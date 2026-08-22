import { chromium, type Page } from "playwright";
import type { Config } from "./config.js";
import { pressEnter } from "./prompt.js";

export interface RegisterResult {
  index: number;
  clientNumber?: string;
  error?: string;
}

/**
 * Opens N browser tabs, fills registration forms, intercepts /api/register
 * responses after user solves captchas and clicks Submit.
 */
export async function registerAll(
  cfg: Config,
  accounts: { index: number; email: string }[],
): Promise<RegisterResult[]> {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();

  const results: RegisterResult[] = accounts.map((a) => ({
    index: a.index,
  }));

  // Phase 1: Open all tabs and wait for forms to load
  const pages: (Page | null)[] = [];
  for (let i = 0; i < accounts.length; i++) {
    const { index, email } = accounts[i];
    try {
      const page = await context.newPage();
      listenForRegisterResponse(page, results, i);
      await page.goto("https://imagekit.io/registration/", {
        waitUntil: "commit",
      });
      await page.waitForSelector('input[placeholder="Enter your work email"]', {
        state: "visible",
        timeout: 60_000,
      });
      pages.push(page);
    } catch (err) {
      results[i].error = (err as Error).message ?? String(err);
      pages.push(null);
      console.log(`  [${index + 1}] ${email} — FAILED to load: ${results[i].error}`);
    }
  }

  // Phase 2: Fill forms (sequential to avoid focus contention)
  for (let i = 0; i < accounts.length; i++) {
    const page = pages[i];
    if (!page) continue;
    const { index, email } = accounts[i];
    try {
      await fillForm(page, cfg, email);
      console.log(`  [${index + 1}] ${email} — form filled`);
    } catch (err) {
      results[i].error = (err as Error).message ?? String(err);
      console.log(`  [${index + 1}] ${email} — FAILED to fill: ${results[i].error}`);
    }
  }

  const filledCount = pages.filter((p) => p !== null).length;
  if (filledCount > 0) {
    console.log(`\n  >>> SOLVE CAPTCHA + click Submit in ${filledCount} browser tab(s) <<<`);
    await pressEnter("Press Enter when ALL submitted...");
    await new Promise((r) => setTimeout(r, 5000));
  }

  await context.close();
  await browser.close();
  return results;
}

async function fillForm(page: Page, cfg: Config, email: string): Promise<void> {
  // Use .fill() directly — works on background tabs without needing focus
  await page.locator('input[placeholder="Enter your work email"]').fill(email);
  await page.locator('input[placeholder="Enter your name"]').fill(cfg.name);
  await page.locator('input[placeholder="Enter a password"]').fill(cfg.password);
}

function listenForRegisterResponse(page: Page, results: RegisterResult[], resultIdx: number): void {
  page.on("response", async (response) => {
    if (!response.url().includes("/api/register")) return;
    try {
      const body = await response.json();
      // Skip the initial captcha-required probe — not the real result
      if (body.isCaptchaRequired) return;
      if (body.success && body.obj?.clientNumber) {
        results[resultIdx].clientNumber = body.obj.clientNumber;
        results[resultIdx].error = undefined;
      } else {
        results[resultIdx].error = JSON.stringify(body);
      }
    } catch {
      // non-JSON response, ignore
    }
  });
}

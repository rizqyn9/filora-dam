import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import { regionCode } from "./config.js";

/**
 * All ImageKit interactions via real browser.
 * Each function takes a Page (owned by the worker's browser context).
 */

// --- Registration ---

export async function register(page: Page, cfg: Config, email: string): Promise<void> {
  await page.goto("https://imagekit.io/registration/", { waitUntil: "commit" });
  await page.waitForSelector('input[placeholder="Enter your work email"]', { state: "visible", timeout: 60_000 });

  await page.locator('input[placeholder="Enter your work email"]').fill(email);
  await page.locator('input[placeholder="Enter your name"]').fill(cfg.name);
  await page.locator('input[placeholder="Enter a password"]').fill(cfg.password);

  // Uncheck marketing checkbox if checked
  const checkbox = page.locator('input[type="checkbox"]');
  if (await checkbox.isChecked()) {
    await checkbox.uncheck();
  }

  // Click submit
  await page.getByRole("button", { name: "Submit" }).click();
}

/**
 * Detect registration outcome after submit.
 * Returns: "email_pending" | "captcha" | "error"
 */
export async function detectRegistrationState(page: Page): Promise<{
  state: "email_pending" | "captcha" | "error";
  error?: string;
}> {
  // Wait for either: redirect to verify page, captcha appearance, or error message
  try {
    await page.waitForFunction(
      () => {
        const url = window.location.href;
        // Success: redirected away from registration or shows verify message
        if (url.includes("verify=true") || url.includes("/login")) return true;
        // Check for visible captcha iframe
        if (document.querySelector('iframe[src*="captcha"], iframe[src*="turnstile"], iframe[src*="recaptcha"]')) return true;
        // Check for error text
        if (document.body.innerText.includes("captcha")) return true;
        if (document.body.innerText.includes("verify")) return true;
        return false;
      },
      { timeout: 15_000 },
    );
  } catch {
    // Timeout — check current state manually
  }

  const url = page.url();

  // Redirected to login with verify=true → email verification pending
  if (url.includes("verify=true") || url.includes("/login")) {
    return { state: "email_pending" };
  }

  // Still on registration page — check for captcha or error
  const bodyText = await page.locator("body").innerText();
  if (bodyText.includes("captcha") || bodyText.includes("Captcha")) {
    return { state: "captcha" };
  }

  // Check if page shows success/verify message
  if (bodyText.includes("verify") || bodyText.includes("Verify")) {
    return { state: "email_pending" };
  }

  return { state: "error", error: `Unexpected state at: ${url}` };
}

// --- Login ---

export async function login(page: Page, cfg: Config, email: string): Promise<void> {
  await page.goto("https://imagekit.io/login/", { waitUntil: "commit" });
  await page.waitForSelector('input[placeholder="Enter your work email"]', { state: "visible", timeout: 60_000 });

  await page.locator('input[placeholder="Enter your work email"]').fill(email);
  await page.locator('input[placeholder="Enter a password"]').fill(cfg.password);
  await page.getByRole("button", { name: "Submit" }).click();
}

/**
 * Detect login outcome.
 * Returns: "authenticated" | "captcha" | "error"
 */
export async function detectLoginState(page: Page): Promise<{
  state: "authenticated" | "captcha" | "error";
  error?: string;
}> {
  try {
    await page.waitForFunction(
      () => {
        const url = window.location.href;
        if (url.includes("/dashboard")) return true;
        if (document.body.innerText.includes("captcha") || document.body.innerText.includes("Captcha")) return true;
        return false;
      },
      { timeout: 15_000 },
    );
  } catch {
    // Timeout
  }

  const url = page.url();
  if (url.includes("/dashboard")) {
    return { state: "authenticated" };
  }

  const bodyText = await page.locator("body").innerText();
  if (bodyText.includes("captcha") || bodyText.includes("Captcha")) {
    return { state: "captcha" };
  }

  return { state: "error", error: `Login stuck at: ${url}` };
}

// --- Onboarding ---

/**
 * Check if onboarding is required by inspecting current URL.
 */
export async function checkOnboardingState(page: Page): Promise<"required" | "completed"> {
  // Navigate to dashboard and see where we end up
  await page.goto("https://imagekit.io/dashboard", { waitUntil: "commit" });
  await page.waitForTimeout(3000);

  const url = page.url();
  if (url.includes("get-started")) return "required";
  return "completed";
}

export async function completeOnboarding(page: Page, cfg: Config): Promise<void> {
  // Ensure we're on get-started
  if (!page.url().includes("get-started")) {
    await page.goto("https://imagekit.io/dashboard/get-started", { waitUntil: "commit" });
  }
  await page.waitForSelector("select", { state: "visible", timeout: 30_000 });

  // 1/3: Role + objective
  await page.locator("select").selectOption({ index: 1 });
  await page.getByText("Real-time image optimization").click();
  await page.getByRole("button", { name: "Next" }).click();
  await page.waitForTimeout(2000);

  // 2/3: Region
  await page.waitForSelector("select", { state: "visible", timeout: 15_000 });
  await page.locator("select").selectOption(regionCode(cfg.region));
  await page.getByRole("button", { name: "Next" }).click();
  await page.waitForTimeout(2000);

  // 3/3: How did you hear
  await page.getByText("Google / Bing").click();
  await page.getByRole("button", { name: "Submit" }).click();

  // Wait for redirect to dashboard
  await page.waitForURL((url) => !url.pathname.includes("get-started"), { timeout: 30_000 });
}

// --- Key Extraction ---

export interface ExtractedKeys {
  publicKey: string;
  imagekitId: string;
  urlEndpoint: string;
}

export async function navigateToApiKeys(page: Page): Promise<void> {
  await page.goto("https://imagekit.io/dashboard/developer/api-keys", { waitUntil: "commit" });
  await page.waitForSelector('input[value^="public_"]', { state: "visible", timeout: 30_000 });
}

export async function extractPublicKeys(page: Page): Promise<ExtractedKeys> {
  const publicKey = await page.locator('input[value^="public_"]').inputValue();

  // Extract URL endpoint from page text
  const endpointText = await page.locator("text=https://ik.imagekit.io/").first().textContent() ?? "";
  const match = endpointText.match(/https:\/\/ik\.imagekit\.io\/\S+/);
  const urlEndpoint = match ? match[0] : "";
  const imagekitId = urlEndpoint.split("/").pop() ?? "";

  return { publicKey, imagekitId, urlEndpoint };
}

export async function clickRevealPrivateKey(page: Page): Promise<void> {
  // Click eye icon next to private key
  await page.locator("td:nth-child(2) .chakra-icon").first().click();
  // Wait for OTP modal
  await page.waitForSelector('[aria-modal="true"]', { state: "visible", timeout: 10_000 });
}

export async function enterOtp(page: Page, code: string): Promise<void> {
  const digits = code.split("");
  const inputs = page.locator('[aria-modal="true"] input');
  for (let i = 0; i < digits.length; i++) {
    await inputs.nth(i).fill(digits[i]);
  }
  await page.getByRole("button", { name: "Verify" }).click();
  await page.waitForTimeout(3000);
}

export async function extractPrivateKey(page: Page): Promise<string> {
  // After OTP verification, private key should be unmasked
  const value = await page.locator('input[value^="private_"]').inputValue();
  if (value.includes("****")) {
    throw new Error("Private key still masked after OTP");
  }
  return value;
}

// --- Utility ---

export async function takeScreenshot(page: Page, label: string): Promise<string> {
  const filename = `output/screenshot_${label}_${Date.now()}.png`;
  await page.screenshot({ path: filename });
  return filename;
}

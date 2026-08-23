import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import { regionCode } from "./config.js";
import { randomDelay, humanType } from "./stealth.js";
import { solveTurnstile } from "./captcha.js";

/**
 * All ImageKit interactions via real browser.
 */

// --- Warm-up ---

/** Visit homepage before target page to build browsing history / cookie reputation. */
async function warmUp(page: Page): Promise<void> {
  await page.goto("https://imagekit.io/", { waitUntil: "load", timeout: 60_000 });
  await page.waitForTimeout(randomDelay(1000, 2000));
}

// --- reCAPTCHA + Buster Auto-Solve ---

/**
 * Solve reCAPTCHA on the current page using the Buster extension.
 *
 * Buster's #solver-button lives inside a CLOSED shadow DOM — can't be clicked
 * via normal selectors. Instead we click #recaptcha-audio-button to switch to
 * audio challenge, then Buster auto-solves it.
 *
 * If that fails, reload and retry up to MAX_ATTEMPTS times.
 */
export async function solveRecaptchaWithBuster(
  page: Page,
  apiEndpoint: string,
): Promise<{ state: "solved" | "error"; error?: string }> {
  const MAX_ATTEMPTS = 3;
  const SOLVE_TIMEOUT = 20_000;

  const anchorFrame = page.frameLocator(
    'iframe[src*="recaptcha/api2/anchor"], iframe[src*="recaptcha/enterprise/anchor"]',
  );
  const bframe = page.frameLocator(
    'iframe[src*="recaptcha/api2/bframe"], iframe[src*="recaptcha/enterprise/bframe"]',
  );

  try {
    // Click the reCAPTCHA checkbox
    await anchorFrame.locator("#recaptcha-anchor-label").click();
    await page.waitForTimeout(randomDelay(1000, 2000));

    // Check if already solved (no challenge needed)
    const alreadySolved = await anchorFrame.locator('[aria-checked="true"]').count();
    if (alreadySolved > 0) {
      return await resubmitAfterCaptcha(page, apiEndpoint);
    }

    // Challenge appeared — try audio solve via Buster
    for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
      // Click audio button to switch to audio challenge (Buster hooks into this)
      const audioBtn = bframe.locator("#recaptcha-audio-button");
      await audioBtn.waitFor({ state: "visible", timeout: 10_000 }).catch(() => {});

      if ((await audioBtn.count()) > 0) {
        await audioBtn.click();
        await page.waitForTimeout(randomDelay(1000, 2000));
      }

      // Now try to click Buster's solver button via the help-button-holder host element
      // The closed shadow root prevents direct access, but we can dispatch click via JS
      await bframe
        .locator(".help-button-holder")
        .evaluate((el) => {
          const shadow = (el as any).shadowRoot;
          if (shadow) {
            const btn = shadow.querySelector("#solver-button");
            if (btn) (btn as HTMLElement).click();
          }
        })
        .catch(() => {
          // Closed shadow root — can't access. Buster may auto-trigger on audio page.
        });

      // Poll for solve
      const solved = await pollForSolve(anchorFrame, SOLVE_TIMEOUT);
      if (solved) {
        await page.waitForTimeout(randomDelay(200, 500));
        return await resubmitAfterCaptcha(page, apiEndpoint);
      }

      // Not solved — reload and retry
      if (attempt < MAX_ATTEMPTS) {
        const reloadBtn = bframe.locator("#recaptcha-reload-button");
        if ((await reloadBtn.count()) > 0) {
          await reloadBtn.click();
          await page.waitForTimeout(randomDelay(1000, 2000));
        }
      }
    }

    return { state: "error", error: `Buster failed after ${MAX_ATTEMPTS} attempts` };
  } catch (err) {
    return {
      state: "error",
      error: `reCAPTCHA solve error: ${err instanceof Error ? err.message : String(err)}`,
    };
  }
}

/** Poll the anchor frame for solved state. */
async function pollForSolve(
  anchorFrame: ReturnType<Page["frameLocator"]>,
  timeoutMs: number,
): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const solved = await anchorFrame.locator('[aria-checked="true"]').count();
    if (solved > 0) return true;
    await new Promise((r) => setTimeout(r, 500));
  }
  return false;
}

/** After captcha solved, re-click Submit and wait for API response. */
async function resubmitAfterCaptcha(
  page: Page,
  apiEndpoint: string,
): Promise<{ state: "solved" | "error"; error?: string }> {
  const responsePromise = page.waitForResponse(
    (res) => res.url().includes(apiEndpoint) && !res.url().includes("csrf"),
    { timeout: 60_000 },
  );

  await page.getByRole("button", { name: "Submit" }).click();

  try {
    const res = await responsePromise;
    const body = await res.json().catch(() => null);

    if (body?.success) return { state: "solved" };
    if (body?.isCaptchaRequired) {
      return { state: "error", error: "Still captcha-required after solving reCAPTCHA" };
    }
    if (body?.error || body?.email) {
      return { state: "error", error: body.email || body.error };
    }
    return { state: "solved" };
  } catch {
    const url = page.url();
    if (url.includes("verify=true") || url.includes("/login") || url.includes("/dashboard")) {
      return { state: "solved" };
    }
    return { state: "error", error: `No API response after reCAPTCHA solve. URL: ${url}` };
  }
}

// --- Turnstile Auto-Solve ---

/**
 * Extract the Turnstile sitekey from the page, solve via CapSolver,
 * inject the token, and re-submit the form.
 * Returns the /api/register (or /api/login) response after re-submission.
 */
export async function autoSolveTurnstile(
  page: Page,
  apiKey: string,
  apiEndpoint: string,
): Promise<{ state: "solved" | "error"; error?: string }> {
  // Extract sitekey from Turnstile iframe or script on page
  const sitekey = await page.evaluate(() => {
    // Method 1: data-sitekey attribute on turnstile div
    const widget = document.querySelector("[data-sitekey]");
    if (widget) return widget.getAttribute("data-sitekey");
    // Method 2: cf-turnstile div
    const cfWidget = document.querySelector(".cf-turnstile");
    if (cfWidget) return cfWidget.getAttribute("data-sitekey");
    // Method 3: iframe src contains sitekey param
    const iframe = document.querySelector('iframe[src*="turnstile"]') as HTMLIFrameElement | null;
    if (iframe) {
      const url = new URL(iframe.src);
      return url.searchParams.get("k") || url.searchParams.get("sitekey");
    }
    // Method 4: search script tags for render() call
    for (const script of document.querySelectorAll("script")) {
      const match = script.textContent?.match(/sitekey['":\s]+['"]([0-9x_A-Za-z-]+)['"]/);
      if (match) return match[1];
    }
    return null;
  });

  if (!sitekey) {
    return { state: "error", error: "Could not extract Turnstile sitekey from page" };
  }

  // Solve via CapSolver
  const result = await solveTurnstile(apiKey, page.url(), sitekey);
  if (!result.success || !result.token) {
    return { state: "error", error: result.error ?? "Solver returned no token" };
  }

  // Inject token into page — Turnstile stores it in a hidden input or callback
  await page.evaluate((token) => {
    // Set hidden input value (standard Turnstile integration)
    const input = document.querySelector(
      'input[name="cf-turnstile-response"], input[name="g-recaptcha-response"]',
    ) as HTMLInputElement | null;
    if (input) {
      input.value = token;
      input.dispatchEvent(new Event("input", { bubbles: true }));
    }
    // Also try the turnstile callback if exposed on window
    const win = window as any;
    if (typeof win.turnstileCallback === "function") {
      win.turnstileCallback(token);
    }
    // Try setting on the cf-turnstile widget's response property
    const widget = document.querySelector(".cf-turnstile") as any;
    if (widget) {
      widget.dataset.response = token;
    }
  }, result.token);

  await page.waitForTimeout(randomDelay(200, 500));

  // Listen for the API response after re-submitting
  const responsePromise = page.waitForResponse(
    (res) => res.url().includes(apiEndpoint) && !res.url().includes("csrf"),
    { timeout: 60_000 },
  );

  // Re-click submit
  await page.getByRole("button", { name: "Submit" }).click();

  try {
    const res = await responsePromise;
    const body = await res.json().catch(() => null);

    if (body?.success) return { state: "solved" };
    if (body?.isCaptchaRequired) {
      return { state: "error", error: "Still captcha-required after token injection" };
    }
    if (body?.error || body?.email) {
      return { state: "error", error: body.email || body.error };
    }
    return { state: "solved" };
  } catch {
    // Check if page navigated (success)
    const url = page.url();
    if (url.includes("verify=true") || url.includes("/login") || url.includes("/dashboard")) {
      return { state: "solved" };
    }
    return { state: "error", error: `No API response after token inject. URL: ${url}` };
  }
}

// --- Registration ---

/**
 * Fill and submit registration form.
 * Returns the result by intercepting the /api/register response.
 */
export async function register(
  page: Page,
  cfg: Config,
  email: string,
): Promise<{
  state: "email_pending" | "captcha" | "error";
  error?: string;
}> {
  await warmUp(page);
  await page.goto("https://imagekit.io/registration/", { waitUntil: "load", timeout: 120_000 });
  await page.waitForSelector('input[placeholder="Enter your work email"]', {
    state: "visible",
    timeout: 60_000,
  });
  await page.waitForTimeout(randomDelay(1500, 3000));

  // Fill form with human-like typing
  await humanType(page, 'input[placeholder="Enter your work email"]', email);
  await page.waitForTimeout(randomDelay(300, 800));
  await humanType(page, 'input[placeholder="Enter your name"]', cfg.name);
  await page.waitForTimeout(randomDelay(300, 800));
  await humanType(page, 'input[placeholder="Enter a password"]', cfg.password);

  // Uncheck marketing checkbox if checked
  const checkbox = page.locator('input[type="checkbox"]');
  if (await checkbox.isChecked()) {
    await checkbox.uncheck();
  }

  await page.waitForTimeout(randomDelay(400, 1000));

  // Listen for /api/register response BEFORE clicking submit
  const responsePromise = page.waitForResponse(
    (res) => res.url().includes("/api/register") && !res.url().includes("csrf"),
    { timeout: 60_000 },
  );

  // Click submit
  await page.getByRole("button", { name: "Submit" }).click();

  // Wait for the API response
  try {
    const res = await responsePromise;
    const body = await res.json().catch(() => null);

    if (!body) {
      return { state: "error", error: `Non-JSON response (${res.status()})` };
    }

    // Captcha required — first attempt without captcha was rejected
    if (body.isCaptchaRequired) {
      return { state: "captcha" };
    }

    // Success
    if (body.success) {
      return { state: "email_pending" };
    }

    // Error (e.g. email exists)
    return { state: "error", error: body.email || body.error || JSON.stringify(body) };
  } catch {
    // Timeout waiting for response — check page state
    const url = page.url();
    if (url.includes("verify=true") || url.includes("/login")) {
      return { state: "email_pending" };
    }
    return { state: "error", error: `No /api/register response. URL: ${url}` };
  }
}

/**
 * After captcha is solved by human, wait for the second /api/register response.
 * NOTE: Call this BEFORE prompting the human — it sets up the listener.
 */
export function listenForRegistrationSuccess(page: Page): Promise<{
  state: "email_pending" | "error";
  error?: string;
}> {
  return page
    .waitForResponse(
      (r) => {
        if (!r.url().includes("/api/register") || r.url().includes("csrf")) return false;
        // Only match successful responses (ignore the captcha-required ones)
        return r.status() === 200;
      },
      { timeout: 180_000 },
    )
    .then(async (res) => {
      const body = await res.json().catch(() => null);
      if (body?.success) return { state: "email_pending" as const };
      if (body?.isCaptchaRequired) return { state: "error" as const, error: "Still needs captcha" };
      return { state: "error" as const, error: body?.email || JSON.stringify(body) };
    })
    .catch(() => {
      const url = page.url();
      if (url.includes("verify=true") || url.includes("/login")) {
        return { state: "email_pending" as const };
      }
      return {
        state: "error" as const,
        error: `Timeout waiting for register after captcha. URL: ${url}`,
      };
    });
}

// --- Login ---

export async function login(
  page: Page,
  cfg: Config,
  email: string,
): Promise<{
  state: "authenticated" | "captcha" | "error";
  error?: string;
}> {
  await warmUp(page);
  await page.goto("https://imagekit.io/login/", { waitUntil: "load", timeout: 120_000 });
  await page.waitForSelector('input[placeholder="Enter your work email"]', {
    state: "visible",
    timeout: 60_000,
  });
  await page.waitForTimeout(randomDelay(1500, 3000));

  await humanType(page, 'input[placeholder="Enter your work email"]', email);
  await page.waitForTimeout(randomDelay(300, 800));
  await humanType(page, 'input[placeholder="Enter a password"]', cfg.password);
  await page.waitForTimeout(randomDelay(400, 1000));

  // Listen for /api/login response before clicking
  const responsePromise = page.waitForResponse(
    (res) => res.url().includes("/api/login") && !res.url().includes("csrf"),
    { timeout: 60_000 },
  );

  await page.getByRole("button", { name: "Submit" }).click();

  try {
    const res = await responsePromise;
    const body = await res.json().catch(() => null);

    if (body?.isCaptchaRequired) {
      return { state: "captcha" };
    }

    if (res.status() >= 400 || body?.error) {
      return { state: "error", error: body?.error || `status ${res.status()}` };
    }

    // Wait for redirect to dashboard
    await page
      .waitForURL((url) => url.pathname.includes("/dashboard"), { timeout: 30_000 })
      .catch(() => {});
    return { state: "authenticated" };
  } catch {
    // Check if we landed on dashboard anyway
    if (page.url().includes("/dashboard")) {
      return { state: "authenticated" };
    }
    // Check for captcha on page (no API call was made — captcha blocked client-side)
    const hasIframe = await page
      .locator('iframe[src*="captcha"], iframe[src*="turnstile"], iframe[src*="challenge"]')
      .count();
    if (hasIframe > 0) {
      return { state: "captcha" };
    }
    return { state: "error", error: `Login stuck at: ${page.url()}` };
  }
}

export async function detectLoginState(page: Page): Promise<{
  state: "authenticated" | "captcha" | "error";
  error?: string;
}> {
  try {
    await page.waitForURL((url) => url.pathname.includes("/dashboard"), { timeout: 30_000 });
    return { state: "authenticated" };
  } catch {
    const bodyText = await page.locator("body").innerText();
    if (bodyText.includes("captcha") || bodyText.includes("Captcha")) {
      return { state: "captcha" };
    }
    return { state: "error", error: `Login stuck at: ${page.url()}` };
  }
}

/**
 * After login captcha is solved by human, wait for the successful /api/login response.
 * Call BEFORE prompting the human so the listener is active when they submit.
 */
export function listenForLoginSuccess(page: Page): Promise<{
  state: "authenticated" | "error";
  error?: string;
}> {
  return page
    .waitForResponse(
      (r) => {
        if (!r.url().includes("/api/login") || r.url().includes("csrf")) return false;
        return r.status() === 200;
      },
      { timeout: 180_000 },
    )
    .then(async () => {
      await page
        .waitForURL((url) => url.pathname.includes("/dashboard"), { timeout: 30_000 })
        .catch(() => {});
      if (page.url().includes("/dashboard")) return { state: "authenticated" as const };
      return { state: "error" as const, error: "Login OK but no dashboard redirect" };
    })
    .catch(() => {
      if (page.url().includes("/dashboard")) return { state: "authenticated" as const };
      return { state: "error" as const, error: `Timeout after login captcha. URL: ${page.url()}` };
    });
}

// --- Onboarding ---

export async function checkOnboardingState(page: Page): Promise<"required" | "completed"> {
  await page.goto("https://imagekit.io/dashboard", { waitUntil: "commit" });
  await page.waitForTimeout(3000);
  return page.url().includes("get-started") ? "required" : "completed";
}

export async function completeOnboarding(page: Page, cfg: Config): Promise<void> {
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
  const endpointText =
    (await page.locator("text=https://ik.imagekit.io/").first().textContent()) ?? "";
  const match = endpointText.match(/https:\/\/ik\.imagekit\.io\/\S+/);
  const urlEndpoint = match ? match[0] : "";
  const imagekitId = urlEndpoint.split("/").pop() ?? "";
  return { publicKey, imagekitId, urlEndpoint };
}

export async function clickRevealPrivateKey(page: Page): Promise<void> {
  await page.locator("td:nth-child(2) .chakra-icon").first().click();
  await page.waitForSelector('[aria-modal="true"]', { state: "visible", timeout: 10_000 });
}

export async function enterOtp(page: Page, code: string): Promise<void> {
  const digits = code.split("");

  // Fill each pin input via JS — set value + trigger React's onChange
  for (let i = 0; i < digits.length; i++) {
    await page.evaluate(
      ({ index, digit }) => {
        const input = document.querySelector(`input[data-index="${index}"]`) as HTMLInputElement;
        if (!input) return;
        // Focus + set value + dispatch events that React/Chakra listens to
        input.focus();
        const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
          window.HTMLInputElement.prototype,
          "value",
        )!.set!;
        nativeInputValueSetter.call(input, digit);
        input.dispatchEvent(new Event("input", { bubbles: true }));
        input.dispatchEvent(new Event("change", { bubbles: true }));
      },
      { index: i, digit: digits[i] },
    );
    await page.waitForTimeout(100);
  }

  await page.waitForTimeout(1000);

  // Check if modal auto-closed (ImageKit auto-verifies on last digit)
  const modalVisible = (await page.locator('[aria-modal="true"]').count()) > 0;
  if (!modalVisible) return;

  // Try clicking Verify if enabled
  const verifyBtn = page.getByRole("button", { name: "Verify" });
  const isDisabled = await verifyBtn.isDisabled().catch(() => true);
  if (!isDisabled) {
    await verifyBtn.click();
    await page.waitForTimeout(3000);
    return;
  }

  // Button disabled — nativeInputValueSetter didn't trigger Chakra properly
  // Fallback: click first input and type via keyboard
  const firstInput = page.locator('[aria-modal="true"] input').first();
  await firstInput.click();
  // Clear all inputs first
  for (let i = 0; i < 6; i++) {
    await page.keyboard.press("Backspace");
  }
  await page.keyboard.type(code, { delay: 150 });
  await page.waitForTimeout(2000);

  // Check if modal closed after keyboard input
  const stillVisible = (await page.locator('[aria-modal="true"]').count()) > 0;
  if (!stillVisible) return;

  // Try Verify button again
  await verifyBtn.click({ timeout: 10_000 }).catch(() => {});
  await page.waitForTimeout(3000);
}

export async function extractPrivateKey(page: Page): Promise<string> {
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

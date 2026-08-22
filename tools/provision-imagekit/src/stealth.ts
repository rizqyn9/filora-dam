/**
 * Anti-detection config for Playwright persistent contexts.
 * Reduces captcha trigger rate on ImageKit (Cloudflare Turnstile).
 */

/** Chrome launch args that suppress automation signals. */
export const BROWSER_ARGS = [
  "--disable-blink-features=AutomationControlled",
  "--no-first-run",
  "--no-default-browser-check",
];

/** Context options for realistic fingerprint. */
export const CONTEXT_OPTIONS = {
  viewport: { width: 1280, height: 800 },
  locale: "en-US",
  timezoneId: "Asia/Jakarta",
  userAgent: undefined, // use default Chromium UA — custom UA is more suspicious
} as const;

/** Random delay between min and max ms. */
export function randomDelay(minMs: number, maxMs: number): number {
  return minMs + Math.random() * (maxMs - minMs);
}

/** Human-like typing: click field then type with random per-key delay. */
export async function humanType(
  page: import("playwright").Page,
  selector: string,
  text: string,
): Promise<void> {
  await page.locator(selector).click();
  await page.waitForTimeout(randomDelay(200, 500));
  await page.keyboard.type(text, { delay: randomDelay(40, 100) });
}

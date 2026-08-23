/**
 * Pre-configure Buster extension storage with Google Cloud Speech-to-Text API key.
 * Must be called once per persistent context after launch.
 */

import type { BrowserContext } from "playwright";

/**
 * Set Buster's storage via service worker or extension options page.
 * Buster uses chrome.storage.local with these keys:
 *   speechService: "googleSpeechApi"
 *   googleSpeechApiKey: "<api-key>"
 */
export async function configureBuster(ctx: BrowserContext, apiKey: string): Promise<boolean> {
  try {
    // Get service worker for the extension
    let sw = ctx.serviceWorkers().find((w) => w.url().includes("background"));

    if (!sw) {
      sw = (await ctx.waitForEvent("serviceworker", { timeout: 10_000 }).catch(() => null)) as any;
    }

    if (!sw) {
      // Fallback: open extension options page to set storage
      const workers = ctx.serviceWorkers();
      let extId: string | null = null;

      for (const w of workers) {
        const match = w.url().match(/chrome-extension:\/\/([^/]+)/);
        if (match) {
          extId = match[1];
          break;
        }
      }

      if (!extId) return false;

      const extPage = await ctx.newPage();
      await extPage.goto(`chrome-extension://${extId}/src/options/index.html`, {
        waitUntil: "load",
        timeout: 10_000,
      });

      await extPage.evaluate((key) => {
        return (globalThis as any).chrome.storage.local.set({
          speechService: "googleSpeechApi",
          googleSpeechApiKey: key,
        });
      }, apiKey);

      await extPage.close();
      return true;
    }

    // Use service worker to set storage directly
    await sw.evaluate((key) => {
      return (globalThis as any).chrome.storage.local.set({
        speechService: "googleSpeechApi",
        googleSpeechApiKey: key,
      });
    }, apiKey);

    return true;
  } catch (err) {
    console.error("Failed to configure Buster:", err instanceof Error ? err.message : err);
    return false;
  }
}

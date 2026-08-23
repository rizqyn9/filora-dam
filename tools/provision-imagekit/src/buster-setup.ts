/**
 * Pre-configure Buster extension storage with IBM Watson STT credentials.
 * Must be called once per persistent context after launch.
 */

import type { BrowserContext } from "playwright";

export interface BusterConfig {
  watsonSttUrl: string;
  watsonSttApiKey: string;
}

/**
 * Set Buster's storage via service worker.
 * Buster uses chrome.storage.local with these keys:
 *   speechService: "ibmSpeechApi"
 *   ibmSpeechApiUrl: "https://api.<region>.speech-to-text.watson.cloud.ibm.com/..."
 *   ibmSpeechApiKey: "<api-key>"
 */
export async function configureBuster(ctx: BrowserContext, cfg: BusterConfig): Promise<boolean> {
  try {
    // Get service worker for the extension
    let sw = ctx.serviceWorkers().find((w) => w.url().includes("background"));

    if (!sw) {
      // Wait for the extension service worker to start
      sw = (await ctx.waitForEvent("serviceworker", { timeout: 10_000 }).catch(() => null)) as any;
    }

    if (!sw) {
      // Fallback: open a page in the extension context to trigger storage set
      const extPage = await ctx.newPage();
      await extPage
        .goto("chrome://extensions/", { waitUntil: "load", timeout: 10_000 })
        .catch(() => {});

      // Use the page to evaluate chrome.storage via a content script workaround
      // Since we can't directly access chrome.storage from a normal page,
      // navigate to the extension's options page
      const pages = ctx.pages();
      // Try to find extension ID from service workers or background pages
      const workers = ctx.serviceWorkers();
      let extId: string | null = null;

      for (const w of workers) {
        const match = w.url().match(/chrome-extension:\/\/([^/]+)/);
        if (match) {
          extId = match[1];
          break;
        }
      }

      if (!extId) {
        await extPage.close();
        return false;
      }

      await extPage.goto(`chrome-extension://${extId}/src/options/index.html`, {
        waitUntil: "load",
        timeout: 10_000,
      });

      await extPage.evaluate(
        ({ url, key }) => {
          return (globalThis as any).chrome.storage.local.set({
            speechService: "ibmSpeechApi",
            ibmSpeechApiUrl: url,
            ibmSpeechApiKey: key,
          });
        },
        { url: cfg.watsonSttUrl, key: cfg.watsonSttApiKey },
      );

      await extPage.close();
      return true;
    }

    // Use service worker to set storage directly
    await sw.evaluate(
      ({ url, key }) => {
        return (globalThis as any).chrome.storage.local.set({
          speechService: "ibmSpeechApi",
          ibmSpeechApiUrl: url,
          ibmSpeechApiKey: key,
        });
      },
      { url: cfg.watsonSttUrl, key: cfg.watsonSttApiKey },
    );

    return true;
  } catch (err) {
    console.error("Failed to configure Buster:", err instanceof Error ? err.message : err);
    return false;
  }
}

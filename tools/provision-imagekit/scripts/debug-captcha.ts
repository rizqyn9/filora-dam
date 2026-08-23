/**
 * Debug script: open ImageKit registration, trigger captcha, inspect selectors.
 * Run: cd tools/provision-imagekit && bun scripts/debug-captcha.ts
 */

import { chromium } from "playwright";
import { resolve } from "path";

const BUSTER_PATH = resolve(import.meta.dirname, "../extensions/buster");

async function main() {
  const ctx = await chromium.launchPersistentContext("/tmp/debug-captcha", {
    headless: false,
    args: [
      "--disable-blink-features=AutomationControlled",
      `--disable-extensions-except=${BUSTER_PATH}`,
      `--load-extension=${BUSTER_PATH}`,
    ],
    viewport: { width: 1280, height: 800 },
  });

  const page = await ctx.newPage();
  await page.goto("https://imagekit.io/registration/", { waitUntil: "load" });
  await page.waitForSelector('input[placeholder="Enter your work email"]', { state: "visible" });

  // Fill form
  await page.locator('input[placeholder="Enter your work email"]').fill("debug@dam.ke-gap-bocil.my.id");
  await page.locator('input[placeholder="Enter your name"]').fill("Debug Test");
  await page.locator('input[placeholder="Enter a password"]').fill("MyStr0ngPass!");

  // Uncheck marketing
  const checkbox = page.locator('input[type="checkbox"]');
  if (await checkbox.isChecked()) await checkbox.uncheck();

  // Submit to trigger captcha
  await page.getByRole("button", { name: "Submit" }).click();
  console.log("Submitted — waiting for captcha...");

  await page.waitForTimeout(5000);

  // Check for reCAPTCHA anchor iframe
  const anchorIframe = page.locator('iframe[src*="recaptcha/api2/anchor"], iframe[src*="recaptcha/enterprise/anchor"]');
  const anchorCount = await anchorIframe.count();
  console.log(`Anchor iframes found: ${anchorCount}`);

  if (anchorCount > 0) {
    const anchorFrame = page.frameLocator('iframe[src*="recaptcha/api2/anchor"], iframe[src*="recaptcha/enterprise/anchor"]');

    // Try #recaptcha-anchor-label
    const labelCount = await anchorFrame.locator("#recaptcha-anchor-label").count();
    console.log(`#recaptcha-anchor-label found: ${labelCount}`);

    // Try clicking it
    if (labelCount > 0) {
      console.log("Clicking #recaptcha-anchor-label...");
      await anchorFrame.locator("#recaptcha-anchor-label").click();
      await page.waitForTimeout(3000);

      // Check for bframe (challenge popup)
      const bframeIframe = page.locator('iframe[src*="recaptcha/api2/bframe"], iframe[src*="recaptcha/enterprise/bframe"]');
      const bframeCount = await bframeIframe.count();
      console.log(`Bframe iframes found: ${bframeCount}`);

      if (bframeCount > 0) {
        const bframe = page.frameLocator('iframe[src*="recaptcha/api2/bframe"], iframe[src*="recaptcha/enterprise/bframe"]');

        // Check for #solver-button (Buster injects this)
        await page.waitForTimeout(3000); // give Buster time to inject
        const solverCount = await bframe.locator("#solver-button").count();
        console.log(`#solver-button found: ${solverCount}`);

        if (solverCount > 0) {
          console.log("Clicking #solver-button...");
          await bframe.locator("#solver-button").click();
          console.log("Buster triggered! Watch the browser...");
          await page.waitForTimeout(30000);
        } else {
          console.log("Buster button not found. Dumping bframe content...");
          const html = await bframe.locator("body").innerHTML();
          console.log(html.slice(0, 2000));
        }
      }
    }
  } else {
    console.log("No reCAPTCHA iframe found. Checking page state...");
    console.log("URL:", page.url());
    const text = await page.locator("body").innerText();
    console.log("Body text (first 500):", text.slice(0, 500));
  }

  console.log("\nBrowser stays open. Press Ctrl+C to exit.");
  await new Promise(() => {}); // keep alive
}

main().catch(console.error);

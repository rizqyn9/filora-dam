import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import type { Batch, Status } from "./state.js";
import { transition, fail, update } from "./state.js";
import * as browser from "./browser.js";

// @ts-ignore — gmail is JS, no types
import { waitForMessage } from "./gmail/gmail.js";

export interface WorkerCallbacks {
  /** Called when account needs human intervention. Worker blocks until resolved. */
  onHumanRequired(accountId: string, email: string, reason: string): Promise<void>;
  /** Log a message for this account. */
  log(accountId: string, msg: string): void;
}

/**
 * Runs a single account through the full state machine.
 * Owns a Page within the shared BrowserContext.
 * Blocks on CAPTCHA until operator resolves.
 */
export async function runAccount(
  ctx: BrowserContext,
  cfg: Config,
  batch: Batch,
  idx: number,
  cb: WorkerCallbacks,
): Promise<void> {
  const account = batch.accounts[idx];
  const { id, email } = account;
  let page: Page | null = null;

  try {
    page = await ctx.newPage();

    // --- REGISTRATION ---
    if (account.status === "CREATED") {
      await transition(batch, idx, "REGISTRATION");
      cb.log(id, "registering...");

      await browser.register(page, cfg, email);
      const regState = await browser.detectRegistrationState(page);

      if (regState.state === "captcha") {
        await transition(batch, idx, "REGISTRATION_HUMAN_REQUIRED");
        cb.log(id, "CAPTCHA — solve in browser");
        await cb.onHumanRequired(id, email, "registration captcha");
        // Re-check after human intervention
        const afterHuman = await browser.detectRegistrationState(page);
        if (afterHuman.state !== "email_pending") {
          await fail(batch, idx, `Post-captcha state: ${afterHuman.state} ${afterHuman.error ?? ""}`);
          return;
        }
      } else if (regState.state === "error") {
        await fail(batch, idx, regState.error ?? "Registration failed");
        await browser.takeScreenshot(page, `${id}_reg_error`);
        return;
      }

      await transition(batch, idx, "EMAIL_VERIFICATION_PENDING");
      cb.log(id, "registered — waiting for verification email");
    }

    // --- EMAIL VERIFICATION ---
    if (account.status === "EMAIL_VERIFICATION_PENDING") {
      cb.log(id, "polling Gmail for verification email...");

      const msg = await waitForMessage({
        query: `to:${email} from:imagekit subject:"Account Verification" newer_than:10m`,
        timeoutMs: 180_000,
        intervalMs: 3000,
      });

      // Extract verification URL from HTML body
      const verifyUrl = extractVerificationUrl(msg.html);
      if (!verifyUrl) {
        await fail(batch, idx, "Could not extract verification URL from email");
        return;
      }

      // Navigate to verification URL
      await page.goto(verifyUrl, { waitUntil: "commit" });
      await page.waitForTimeout(3000);

      await transition(batch, idx, "EMAIL_VERIFIED");
      cb.log(id, "email verified");
    }

    // --- LOGIN ---
    if (account.status === "EMAIL_VERIFIED") {
      await transition(batch, idx, "LOGIN");
      cb.log(id, "logging in...");

      await browser.login(page, cfg, email);
      const loginState = await browser.detectLoginState(page);

      if (loginState.state === "captcha") {
        await transition(batch, idx, "LOGIN_HUMAN_REQUIRED");
        cb.log(id, "LOGIN CAPTCHA — solve in browser");
        await cb.onHumanRequired(id, email, "login captcha");
        // Re-check
        const afterHuman = await browser.detectLoginState(page);
        if (afterHuman.state !== "authenticated") {
          await fail(batch, idx, `Post-login-captcha state: ${afterHuman.state}`);
          return;
        }
      } else if (loginState.state === "error") {
        await fail(batch, idx, loginState.error ?? "Login failed");
        await browser.takeScreenshot(page, `${id}_login_error`);
        return;
      }

      await transition(batch, idx, "AUTHENTICATED");
      cb.log(id, "authenticated");
    }

    // --- ONBOARDING ---
    if (account.status === "AUTHENTICATED") {
      const onboardState = await browser.checkOnboardingState(page);
      if (onboardState === "required") {
        await transition(batch, idx, "ONBOARDING_REQUIRED");
        cb.log(id, "completing onboarding...");
        await browser.completeOnboarding(page, cfg);
      }
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      cb.log(id, "onboarded");
    }

    // Resume from ONBOARDING_REQUIRED (e.g. on resume after crash)
    if (account.status === "ONBOARDING_REQUIRED") {
      cb.log(id, "completing onboarding...");
      await browser.completeOnboarding(page, cfg);
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      cb.log(id, "onboarded");
    }

    // --- CONFIG RETRIEVAL ---
    if (account.status === "ONBOARDING_COMPLETED") {
      await transition(batch, idx, "CONFIG_RETRIEVAL");
      cb.log(id, "extracting keys...");

      await browser.navigateToApiKeys(page);
      const keys = await browser.extractPublicKeys(page);
      await update(batch, idx, {
        public_key: keys.publicKey,
        imagekit_id: keys.imagekitId,
        url_endpoint: keys.urlEndpoint,
      });
      cb.log(id, `public: ${keys.publicKey}`);

      // Initiate private key reveal
      await browser.clickRevealPrivateKey(page);
      await transition(batch, idx, "PRIVATE_KEY_OTP_PENDING");
      cb.log(id, "OTP sent — polling Gmail...");
    }

    // --- PRIVATE KEY OTP ---
    if (account.status === "PRIVATE_KEY_OTP_PENDING") {
      // Wait for OTP email
      const otpMsg = await waitForMessage({
        query: `to:${email} from:imagekit subject:"Verification code for a sensitive action" newer_than:5m`,
        timeoutMs: 120_000,
        intervalMs: 3000,
      });

      const otp = extractOtp(otpMsg.text || otpMsg.html);
      if (!otp) {
        await fail(batch, idx, "Could not extract OTP from email");
        return;
      }
      cb.log(id, "OTP received, entering...");

      await browser.enterOtp(page, otp);
      const privateKey = await browser.extractPrivateKey(page);
      await update(batch, idx, { private_key: privateKey });
      await transition(batch, idx, "PRIVATE_KEY_REVEALED");
      cb.log(id, "private key revealed");
    }

    // --- COMPLETED ---
    if (account.status === "PRIVATE_KEY_REVEALED") {
      await transition(batch, idx, "COMPLETED");
      cb.log(id, "COMPLETED");
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    await fail(batch, idx, msg);
    cb.log(id, `FAILED: ${msg}`);
    if (page) {
      await browser.takeScreenshot(page, `${id}_crash`).catch(() => {});
    }
  } finally {
    if (page) await page.close().catch(() => {});
  }
}

// --- Helpers ---

function extractVerificationUrl(html: string): string | null {
  // Look for href in an <a> tag containing "Verify" text or imagekit verify URL
  const hrefMatch = html.match(/href="(https:\/\/imagekit\.io\/[^"]*verify[^"]*)"/i);
  if (hrefMatch) return hrefMatch[1];
  // Fallback: any imagekit URL with token-like params
  const urlMatch = html.match(/href="(https:\/\/imagekit\.io\/[^"]+)"/);
  return urlMatch?.[1] ?? null;
}

function extractOtp(text: string): string | null {
  // Find a 6-digit number
  const match = text.match(/\b(\d{6})\b/);
  return match?.[1] ?? null;
}

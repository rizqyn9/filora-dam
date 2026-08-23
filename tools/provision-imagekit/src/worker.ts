import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import type { Batch } from "./state.js";
import { transition, fail, update } from "./state.js";
import { createAccountLogger } from "./logger.js";
import * as browser from "./browser.js";

// @ts-ignore — gmail is JS, no types
import { waitForMessage } from "./gmail/gmail.js";

export interface WorkerCallbacks {
  onHumanRequired?(accountId: string, email: string, reason: string): Promise<void>;
}

export async function runAccount(
  ctx: BrowserContext,
  cfg: Config,
  batch: Batch,
  idx: number,
  cb: WorkerCallbacks,
): Promise<void> {
  const account = batch.accounts[idx];
  const { id, email } = account;

  const log = createAccountLogger(id, email);

  let page: Page | null = null;
  const startTime = Date.now();

  log.info(`start | ${email} | state=${account.status}`);

  try {
    page = await ctx.newPage();

    // Resume login for accounts past registration
    const needsLoginForResume = [
      "AUTHENTICATED",
      "ONBOARDING_REQUIRED",
      "ONBOARDING_COMPLETED",
      "CONFIG_RETRIEVAL",
      "PRIVATE_KEY_OTP_PENDING",
    ].includes(account.status);

    if (needsLoginForResume) {
      log.info("resume — logging in...");
      const loginResult = await browser.login(page, cfg, email);
      if (loginResult.state === "captcha") {
        log.warn("login captcha on resume");
        const successPromise = browser.listenForLoginSuccess(page);
        await cb.onHumanRequired?.(id, email, "login captcha (resume)");
        const afterHuman = await successPromise;
        if (afterHuman.state !== "authenticated") {
          await fail(batch, idx, afterHuman.error ?? "Login failed on resume");
          return;
        }
      } else if (loginResult.state === "error") {
        await fail(batch, idx, loginResult.error ?? "Login failed on resume");
        return;
      }
      log.success(`resume login OK | url=${page.url()}`);
    }

    // --- REGISTRATION ---
    if (account.status === "CREATED") {
      await transition(batch, idx, "REGISTRATION");
      log.state("CREATED", "REGISTRATION");

      const regResult = await browser.register(page, cfg, email);
      log.debug(
        `register response: ${regResult.state}${regResult.error ? ` | ${regResult.error}` : ""}`,
      );

      if (regResult.state === "captcha") {
        await transition(batch, idx, "REGISTRATION_HUMAN_REQUIRED");
        log.warn("CAPTCHA — solve in browser and click Submit");
        const successPromise = browser.listenForRegistrationSuccess(page);
        await cb.onHumanRequired?.(id, email, "registration captcha");
        const afterHuman = await successPromise;
        log.debug(`post-captcha: ${afterHuman.state}`);
        if (afterHuman.state !== "email_pending") {
          await fail(batch, idx, afterHuman.error ?? "Registration failed after captcha");
          return;
        }
      } else if (regResult.state === "error") {
        await fail(batch, idx, regResult.error ?? "Registration failed");
        await browser.takeScreenshot(page, `${id}_reg_error`);
        return;
      }

      await transition(batch, idx, "EMAIL_VERIFICATION_PENDING");
      log.state("REGISTRATION", "EMAIL_VERIFICATION_PENDING");
    }

    // --- EMAIL VERIFICATION ---
    if (account.status === "EMAIL_VERIFICATION_PENDING") {
      const query = `to:${email} from:imagekit subject:"Account Verification" newer_than:10m`;
      log.info("polling Gmail for verification email...");
      log.debug(`query: ${query}`);

      const msg = await waitForMessage({ query, timeoutMs: 180_000, intervalMs: 3000 });
      log.debug(`found: subject="${msg.subject}" date="${msg.date}"`);

      const verifyUrl = extractVerificationUrl(msg.html);
      if (!verifyUrl) {
        await fail(batch, idx, "Could not extract verification URL from email");
        return;
      }
      log.debug(`verify url: ${verifyUrl.slice(0, 80)}...`);

      await page.goto(verifyUrl, { waitUntil: "commit" });
      await page.waitForTimeout(3000);

      await transition(batch, idx, "EMAIL_VERIFIED");
      log.state("EMAIL_VERIFICATION_PENDING", "EMAIL_VERIFIED");
    }

    // --- LOGIN ---
    if (account.status === "EMAIL_VERIFIED") {
      await transition(batch, idx, "LOGIN");
      log.state("EMAIL_VERIFIED", "LOGIN");
      log.info("submitting login... (solve captcha if it appears)");

      const loginResult = await browser.login(page, cfg, email);
      log.debug(`login: ${loginResult.state} | url=${page.url()}`);

      if (loginResult.state === "captcha") {
        await transition(batch, idx, "LOGIN_HUMAN_REQUIRED");
        log.warn("LOGIN CAPTCHA — solve in browser");
        const successPromise = browser.listenForLoginSuccess(page);
        await cb.onHumanRequired?.(id, email, "login captcha");
        const afterHuman = await successPromise;
        if (afterHuman.state !== "authenticated") {
          await fail(batch, idx, afterHuman.error ?? "Login failed after captcha");
          return;
        }
      } else if (loginResult.state === "error") {
        await fail(batch, idx, loginResult.error ?? "Login failed");
        await browser.takeScreenshot(page, `${id}_login_error`);
        return;
      }

      await transition(batch, idx, "AUTHENTICATED");
      log.state("LOGIN", "AUTHENTICATED");
    }

    // --- ONBOARDING ---
    if (account.status === "AUTHENTICATED") {
      const onboardState = await browser.checkOnboardingState(page);
      log.debug(`onboarding check: ${onboardState}`);
      if (onboardState === "required") {
        await transition(batch, idx, "ONBOARDING_REQUIRED");
        log.info("completing onboarding...");
        await browser.completeOnboarding(page, cfg);
      }
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      log.state("AUTHENTICATED", "ONBOARDING_COMPLETED");
    }

    if (account.status === "ONBOARDING_REQUIRED") {
      log.info("resuming onboarding...");
      await browser.completeOnboarding(page, cfg);
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      log.state("ONBOARDING_REQUIRED", "ONBOARDING_COMPLETED");
    }

    // --- CONFIG RETRIEVAL ---
    if (account.status === "ONBOARDING_COMPLETED") {
      await transition(batch, idx, "CONFIG_RETRIEVAL");
      log.state("ONBOARDING_COMPLETED", "CONFIG_RETRIEVAL");

      await browser.navigateToApiKeys(page);
      const keys = await browser.extractPublicKeys(page);
      log.info(`public=${keys.publicKey} | id=${keys.imagekitId}`);

      await update(batch, idx, {
        public_key: keys.publicKey,
        imagekit_id: keys.imagekitId,
        url_endpoint: keys.urlEndpoint,
      });

      log.info("clicking reveal private key...");
      await browser.clickRevealPrivateKey(page);
      await transition(batch, idx, "PRIVATE_KEY_OTP_PENDING");
      log.state("CONFIG_RETRIEVAL", "PRIVATE_KEY_OTP_PENDING");
    }

    // --- PRIVATE KEY OTP ---
    if (account.status === "PRIVATE_KEY_OTP_PENDING") {
      const afterEpoch = Math.floor(Date.now() / 1000) - 30;
      const query = `to:${email} from:imagekit subject:"Verification code for a sensitive action" after:${afterEpoch}`;
      log.info("polling Gmail for OTP...");
      log.debug(`query: ${query}`);

      const otpMsg = await waitForMessage({ query, timeoutMs: 120_000, intervalMs: 3000 });
      log.debug(`found: subject="${otpMsg.subject}" date="${otpMsg.date}"`);

      const otpBody = otpMsg.text || otpMsg.html;
      const otp = extractOtp(otpBody);

      if (!otp) {
        log.error(`OTP extraction failed | body_length=${otpBody.length}`);
        log.debug(`body_preview: ${otpBody.slice(0, 200)}`);
        await fail(batch, idx, "Could not extract OTP from email");
        return;
      }

      log.info(`OTP=${otp} | entering...`);
      await browser.enterOtp(page, otp);

      const privateKey = await browser.extractPrivateKey(page);
      log.success(`private key extracted (${privateKey.length} chars)`);

      await update(batch, idx, { private_key: privateKey });
      await transition(batch, idx, "PRIVATE_KEY_REVEALED");
      log.state("PRIVATE_KEY_OTP_PENDING", "PRIVATE_KEY_REVEALED");
    }

    // --- COMPLETED ---
    if (account.status === "PRIVATE_KEY_REVEALED") {
      await transition(batch, idx, "COMPLETED");
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      log.success(`COMPLETED in ${elapsed}s | ${batch.accounts[idx].url_endpoint}`);
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    await fail(batch, idx, msg);
    log.fail(`FAILED after ${elapsed}s: ${msg}`);
    if (page) {
      log.debug(`url: ${page.url()}`);
      const ss = await browser.takeScreenshot(page, `${id}_crash`).catch(() => null);
      if (ss) log.debug(`screenshot: ${ss}`);
    }
  } finally {
    if (page) await page.close().catch(() => {});
  }
}

// --- Helpers ---

function extractVerificationUrl(html: string): string | null {
  const hrefMatch = html.match(/href="(https:\/\/imagekit\.io\/[^"]*verify[^"]*)"/i);
  if (hrefMatch) return hrefMatch[1];
  const urlMatch = html.match(/href="(https:\/\/imagekit\.io\/[^"]+)"/);
  return urlMatch?.[1] ?? null;
}

function extractOtp(text: string): string | null {
  const styledMatch = text.match(/letter-spacing:\s*6px[^>]*>(\d{6})</);
  if (styledMatch) return styledMatch[1];

  const fontMatch = text.match(/font-size:\s*28px[^>]*>(\d{6})</);
  if (fontMatch) return fontMatch[1];

  const tagMatch = text.match(/>(\d{6})<\/div>/);
  if (tagMatch) return tagMatch[1];

  const lineMatch = text.match(/^\s*(\d{6})\s*$/m);
  if (lineMatch) return lineMatch[1];

  const allMatches = [...text.matchAll(/\b(\d{6})\b/g)];
  if (allMatches.length > 0) return allMatches[allMatches.length - 1][1];

  return null;
}

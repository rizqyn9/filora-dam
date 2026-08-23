import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import type { Batch, Status } from "./state.js";
import { transition, fail, update } from "./state.js";
import * as browser from "./browser.js";

// @ts-ignore — gmail is JS, no types
import { waitForMessage } from "./gmail/gmail.js";

export interface WorkerCallbacks {
  onHumanRequired(accountId: string, email: string, reason: string): Promise<void>;
  log(accountId: string, msg: string): void;
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
  let page: Page | null = null;
  const startTime = Date.now();

  const log = (msg: string) => cb.log(id, msg);

  log(`start | email=${email} | state=${account.status}`);

  try {
    page = await ctx.newPage();
    log("browser tab opened");

    // For resumed accounts past login, we need to authenticate first
    const needsLoginForResume = [
      "AUTHENTICATED",
      "ONBOARDING_REQUIRED",
      "ONBOARDING_COMPLETED",
      "CONFIG_RETRIEVAL",
      "PRIVATE_KEY_OTP_PENDING",
    ].includes(account.status);

    if (needsLoginForResume) {
      log("resume — logging in first...");
      const loginResult = await browser.login(page, cfg, email);
      if (loginResult.state === "captcha") {
        if (cfg.capsolver_api_key) {
          log("LOGIN CAPTCHA on resume — auto-solving...");
          const solveResult = await browser.autoSolveTurnstile(
            page,
            cfg.capsolver_api_key,
            "/api/login",
          );
          if (solveResult.state !== "solved") {
            log(`auto-solve failed: ${solveResult.error} — falling back to human`);
            const successPromise = browser.listenForLoginSuccess(page);
            await cb.onHumanRequired(id, email, "login captcha (resume, auto-solve failed)");
            const afterHuman = await successPromise;
            if (afterHuman.state !== "authenticated") {
              await fail(batch, idx, afterHuman.error ?? "Login failed on resume");
              return;
            }
          }
          await page
            .waitForURL((url) => url.pathname.includes("/dashboard"), { timeout: 30_000 })
            .catch(() => {});
        } else {
          log("LOGIN CAPTCHA on resume — solve in browser");
          const successPromise = browser.listenForLoginSuccess(page);
          await cb.onHumanRequired(id, email, "login captcha (resume)");
          const afterHuman = await successPromise;
          if (afterHuman.state !== "authenticated") {
            await fail(batch, idx, afterHuman.error ?? "Login failed on resume");
            return;
          }
        }
      } else if (loginResult.state === "error") {
        await fail(batch, idx, loginResult.error ?? "Login failed on resume");
        return;
      }
      log(`resume login OK | url=${page.url()}`);
    }

    // --- REGISTRATION ---
    if (account.status === "CREATED") {
      await transition(batch, idx, "REGISTRATION");
      log("→ REGISTRATION");

      const regResult = await browser.register(page, cfg, email);
      log(
        `register result: state=${regResult.state}${regResult.error ? ` error=${regResult.error}` : ""}`,
      );

      if (regResult.state === "captcha") {
        // Strategy: Buster (free) → CapSolver (paid) → Human
        let solved = false;

        // 1. Try Buster extension (reCAPTCHA audio + speech recognition)
        if (cfg.google_stt_api_key) {
          log("captcha detected — attempting Buster reCAPTCHA solver...");
          const busterResult = await browser.solveRecaptchaWithBuster(page, "/api/register");
          log(
            `buster result: state=${busterResult.state}${busterResult.error ? ` error=${busterResult.error}` : ""}`,
          );
          if (busterResult.state === "solved") {
            solved = true;
          } else {
            log(`buster failed: ${busterResult.error}`);
          }
        }

        // 2. Fallback: CapSolver API
        if (!solved && cfg.capsolver_api_key) {
          log("trying CapSolver API fallback...");
          const solveResult = await browser.autoSolveTurnstile(
            page,
            cfg.capsolver_api_key,
            "/api/register",
          );
          log(
            `capsolver result: state=${solveResult.state}${solveResult.error ? ` error=${solveResult.error}` : ""}`,
          );
          if (solveResult.state === "solved") {
            solved = true;
          } else {
            log(`capsolver failed: ${solveResult.error}`);
          }
        }

        // 3. Fallback: Human intervention
        if (!solved) {
          await transition(batch, idx, "REGISTRATION_HUMAN_REQUIRED");
          log("→ REGISTRATION_HUMAN_REQUIRED (auto-solve exhausted)");
          const successPromise = browser.listenForRegistrationSuccess(page);
          await cb.onHumanRequired(id, email, "registration captcha (auto-solve failed)");
          log("human resolved, waiting for register response...");
          const afterHuman = await successPromise;
          log(
            `post-captcha result: state=${afterHuman.state}${afterHuman.error ? ` error=${afterHuman.error}` : ""}`,
          );
          if (afterHuman.state !== "email_pending") {
            await fail(batch, idx, afterHuman.error ?? "Registration failed after captcha");
            return;
          }
        }

        await transition(batch, idx, "EMAIL_VERIFICATION_PENDING");
        log("→ EMAIL_VERIFICATION_PENDING");
      } else if (regResult.state === "error") {
        await fail(batch, idx, regResult.error ?? "Registration failed");
        await browser.takeScreenshot(page, `${id}_reg_error`);
        return;
      }

      await transition(batch, idx, "EMAIL_VERIFICATION_PENDING");
      log("→ EMAIL_VERIFICATION_PENDING");
    }

    // --- EMAIL VERIFICATION ---
    if (account.status === "EMAIL_VERIFICATION_PENDING") {
      const query = `to:${email} from:imagekit subject:"Account Verification" newer_than:10m`;
      log(`gmail query: ${query}`);

      const msg = await waitForMessage({ query, timeoutMs: 180_000, intervalMs: 3000 });
      log(`gmail found: subject="${msg.subject}" date="${msg.date}"`);

      const verifyUrl = extractVerificationUrl(msg.html);
      log(`verify URL: ${verifyUrl ?? "NOT FOUND"}`);
      if (!verifyUrl) {
        await fail(batch, idx, "Could not extract verification URL from email");
        return;
      }

      await page.goto(verifyUrl, { waitUntil: "commit" });
      await page.waitForTimeout(3000);
      log(`after verify navigation: url=${page.url()}`);

      await transition(batch, idx, "EMAIL_VERIFIED");
      log("→ EMAIL_VERIFIED");
    }

    // --- LOGIN ---
    if (account.status === "EMAIL_VERIFIED") {
      await transition(batch, idx, "LOGIN");
      log("→ LOGIN | submitting... (solve captcha in browser if it appears)");

      const loginResult = await browser.login(page, cfg, email);
      log(
        `login result: state=${loginResult.state}${loginResult.error ? ` error=${loginResult.error}` : ""}`,
      );

      if (loginResult.state === "captcha") {
        if (cfg.capsolver_api_key) {
          log("login captcha detected — attempting auto-solve via CapSolver...");
          const solveResult = await browser.autoSolveTurnstile(
            page,
            cfg.capsolver_api_key,
            "/api/login",
          );
          log(
            `auto-solve result: state=${solveResult.state}${solveResult.error ? ` error=${solveResult.error}` : ""}`,
          );
          if (solveResult.state !== "solved") {
            log(`auto-solve failed, falling back to human: ${solveResult.error}`);
            await transition(batch, idx, "LOGIN_HUMAN_REQUIRED");
            const successPromise = browser.listenForLoginSuccess(page);
            await cb.onHumanRequired(
              id,
              email,
              `login captcha (auto-solve failed: ${solveResult.error})`,
            );
            const afterHuman = await successPromise;
            if (afterHuman.state !== "authenticated") {
              await fail(batch, idx, afterHuman.error ?? "Login failed after captcha");
              return;
            }
          }
          // If solved, wait for dashboard redirect
          await page
            .waitForURL((url) => url.pathname.includes("/dashboard"), { timeout: 30_000 })
            .catch(() => {});
        } else {
          await transition(batch, idx, "LOGIN_HUMAN_REQUIRED");
          log("→ LOGIN_HUMAN_REQUIRED");
          const successPromise = browser.listenForLoginSuccess(page);
          await cb.onHumanRequired(id, email, "login captcha");
          log("human resolved, waiting for login response...");
          const afterHuman = await successPromise;
          log(
            `post-login-captcha: state=${afterHuman.state}${afterHuman.error ? ` error=${afterHuman.error}` : ""}`,
          );
          if (afterHuman.state !== "authenticated") {
            await fail(batch, idx, afterHuman.error ?? "Login failed after captcha");
            return;
          }
        }
      } else if (loginResult.state === "error") {
        await fail(batch, idx, loginResult.error ?? "Login failed");
        await browser.takeScreenshot(page, `${id}_login_error`);
        return;
      }

      await transition(batch, idx, "AUTHENTICATED");
      log(`→ AUTHENTICATED | url=${page.url()}`);
    }

    // --- ONBOARDING ---
    if (account.status === "AUTHENTICATED") {
      const onboardState = await browser.checkOnboardingState(page);
      log(`onboarding check: ${onboardState} | url=${page.url()}`);
      if (onboardState === "required") {
        await transition(batch, idx, "ONBOARDING_REQUIRED");
        log("→ ONBOARDING_REQUIRED");
        await browser.completeOnboarding(page, cfg);
      }
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      log(`→ ONBOARDING_COMPLETED | url=${page.url()}`);
    }

    if (account.status === "ONBOARDING_REQUIRED") {
      log("resuming onboarding...");
      await browser.completeOnboarding(page, cfg);
      await transition(batch, idx, "ONBOARDING_COMPLETED");
      log(`→ ONBOARDING_COMPLETED | url=${page.url()}`);
    }

    // --- CONFIG RETRIEVAL ---
    if (account.status === "ONBOARDING_COMPLETED") {
      await transition(batch, idx, "CONFIG_RETRIEVAL");
      log("→ CONFIG_RETRIEVAL | navigating to API keys...");

      await browser.navigateToApiKeys(page);
      log(`api keys page loaded | url=${page.url()}`);

      const keys = await browser.extractPublicKeys(page);
      log(
        `extracted: publicKey=${keys.publicKey} imagekitId=${keys.imagekitId} endpoint=${keys.urlEndpoint}`,
      );

      await update(batch, idx, {
        public_key: keys.publicKey,
        imagekit_id: keys.imagekitId,
        url_endpoint: keys.urlEndpoint,
      });

      log("clicking reveal private key...");
      await browser.clickRevealPrivateKey(page);
      await transition(batch, idx, "PRIVATE_KEY_OTP_PENDING");
      log("→ PRIVATE_KEY_OTP_PENDING | OTP modal visible");
    }

    // --- PRIVATE KEY OTP ---
    if (account.status === "PRIVATE_KEY_OTP_PENDING") {
      const afterEpoch = Math.floor(Date.now() / 1000) - 30;
      const query = `to:${email} from:imagekit subject:"Verification code for a sensitive action" after:${afterEpoch}`;
      log(`gmail OTP query: ${query}`);

      const otpMsg = await waitForMessage({ query, timeoutMs: 120_000, intervalMs: 3000 });
      log(`gmail OTP found: subject="${otpMsg.subject}" date="${otpMsg.date}"`);

      const otpBody = otpMsg.text || otpMsg.html;
      const otp = extractOtp(otpBody);
      log(`extracted OTP: ${otp ?? "NOT FOUND"} (body length: ${otpBody.length})`);

      if (!otp) {
        log(`body preview: ${otpBody.slice(0, 300)}`);
        await fail(batch, idx, "Could not extract OTP from email");
        return;
      }

      log(`entering OTP: ${otp}`);
      await browser.enterOtp(page, otp);

      log("checking private key...");
      const privateKey = await browser.extractPrivateKey(page);
      log(`private key extracted (length=${privateKey.length})`);

      await update(batch, idx, { private_key: privateKey });
      await transition(batch, idx, "PRIVATE_KEY_REVEALED");
      log("→ PRIVATE_KEY_REVEALED");
    }

    // --- COMPLETED ---
    if (account.status === "PRIVATE_KEY_REVEALED") {
      await transition(batch, idx, "COMPLETED");
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      log(`✅ COMPLETED in ${elapsed}s`);
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    await fail(batch, idx, msg);
    log(`❌ FAILED after ${elapsed}s: ${msg}`);
    if (page) {
      log(`current url: ${page.url()}`);
      const screenshot = await browser.takeScreenshot(page, `${id}_crash`).catch(() => null);
      if (screenshot) log(`screenshot: ${screenshot}`);
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
  // ImageKit OTP: code in a div with letter-spacing:6px and font-size:28px
  const styledMatch = text.match(/letter-spacing:\s*6px[^>]*>(\d{6})</);
  if (styledMatch) return styledMatch[1];

  const fontMatch = text.match(/font-size:\s*28px[^>]*>(\d{6})</);
  if (fontMatch) return fontMatch[1];

  // 6 digits inside a closing div tag
  const tagMatch = text.match(/>(\d{6})<\/div>/);
  if (tagMatch) return tagMatch[1];

  // Standalone line
  const lineMatch = text.match(/^\s*(\d{6})\s*$/m);
  if (lineMatch) return lineMatch[1];

  // Last 6-digit number in body
  const allMatches = [...text.matchAll(/\b(\d{6})\b/g)];
  if (allMatches.length > 0) return allMatches[allMatches.length - 1][1];

  return null;
}

import { parseArgs } from "util";
import { chromium } from "playwright";
import { loadConfig } from "./config.js";
import { generateEmail } from "./email.js";
import {
  newBatch,
  loadBatch,
  saveBatch,
  batchPath,
  advance,
  fail,
  update,
  indicesWithStatus,
  type Batch,
} from "./state.js";
import { registerAll } from "./register.js";
import { Session } from "./provision.js";
import { ask, pressEnter, closePrompt } from "./prompt.js";
import type { Config } from "./config.js";

const { positionals, values } = parseArgs({
  args: process.argv.slice(2),
  options: {
    config: { type: "string", short: "c", default: "config.yaml" },
    output: { type: "string", short: "o", default: "output" },
  },
  allowPositionals: true,
});

const command = positionals[0] ?? "run";
const cfgPath = values.config!;
const outputDir = values.output!;

async function main() {
  const cfg = loadConfig(cfgPath);

  if (command === "run") {
    const emails = Array.from({ length: cfg.batch_size }, () => generateEmail(cfg.email_domain));
    const batch = newBatch(emails);
    saveBatch(outputDir, batch);
    console.log(`Batch: ${batch.id} (${batch.accounts.length} accounts)`);
    console.log(`State: ${batchPath(outputDir, batch)}\n`);
    await runSteps(cfg, batch);
  } else if (command === "resume") {
    const batchId = positionals[1];
    if (!batchId) {
      console.error("Usage: resume <batch-id>");
      process.exit(1);
    }
    const batch = loadBatch(outputDir, batchId);
    for (const a of batch.accounts) {
      if (a.status === "failed") {
        a.error = undefined;
        if (a.private_key) a.status = "keys_extracted";
        else if (a.public_key) a.status = "onboarded";
        else if (a.client_number) a.status = "registered";
        else a.status = "pending";
      }
    }
    console.log(`Resuming: ${batch.id} (${batch.accounts.length} accounts)`);
    console.log(`State: ${batchPath(outputDir, batch)}\n`);
    await runSteps(cfg, batch);
  } else {
    console.error(`Unknown command: ${command}`);
    process.exit(1);
  }
  closePrompt();
}

async function runSteps(cfg: Config, batch: Batch) {
  // Step 1: Register (headed browser, N tabs)
  const pending = indicesWithStatus(batch, "pending");
  if (pending.length > 0) {
    console.log(`=== STEP 1: Register (${pending.length} accounts) ===`);
    const accounts = pending.map((i) => ({ index: i, email: batch.accounts[i].email }));
    const results = await registerAll(cfg, accounts);
    for (const r of results) {
      if (r.error) {
        fail(batch, r.index, r.error);
        console.log(`  ✗ [${r.index + 1}] ${r.error}`);
      } else {
        update(batch, r.index, { client_number: r.clientNumber });
        advance(batch, r.index, "registered");
        console.log(`  ✓ [${r.index + 1}] ${batch.accounts[r.index].email} → ${r.clientNumber}`);
      }
    }
    saveBatch(outputDir, batch);
    printSummary(batch, "registered");
  }

  // Step 2: Verify emails
  const registered = indicesWithStatus(batch, "registered");

  if (registered.length > 0) {
    console.log(`\n=== STEP 2: Verify emails (${registered.length} accounts) ===`);
    for (const i of registered) console.log(`  - ${batch.accounts[i].email}`);
    await pressEnter("\nClick 'Verify Account' in each email, then press Enter...");
    for (const i of registered) advance(batch, i, "verified");
    saveBatch(outputDir, batch);
  }

  // Steps 3-5: use headless browser for HTTP (ImageKit blocks raw fetch)
  const needsWork =
    indicesWithStatus(batch, "verified").length > 0 ||
    indicesWithStatus(batch, "onboarded").length > 0;
  if (!needsWork) {
    printFinal(batch);
    return;
  }

  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();

  try {
    // Step 3: Login + onboard
    const verified = indicesWithStatus(batch, "verified");
    if (verified.length > 0) {
      console.log(`\n=== STEP 3: Login + Onboard (${verified.length} accounts) ===`);
      for (const i of verified) {
        const email = batch.accounts[i].email;
        const sess = new Session(context);
        try {
          await sess.init();
          await sess.login(cfg, email);
          sess.clientNumber = batch.accounts[i].client_number!;
          await sess.completeOnboarding(cfg);
          advance(batch, i, "onboarded");
          console.log(`  ✓ ${email}`);
        } catch (err) {
          fail(batch, i, String(err));
          console.log(`  ✗ ${email} — ${err}`);
        } finally {
          await sess.dispose();
        }
      }
      saveBatch(outputDir, batch);
      printSummary(batch, "onboarded");
    }

    // Step 4: Get keys + send OTP
    interface ElevEntry {
      idx: number;
      sess: Session;
    }
    const elevated: ElevEntry[] = [];
    const onboarded = indicesWithStatus(batch, "onboarded");
    if (onboarded.length > 0) {
      console.log(`\n=== STEP 4: Extract keys + send OTP (${onboarded.length} accounts) ===`);
      for (const i of onboarded) {
        const email = batch.accounts[i].email;
        const sess = new Session(context);
        try {
          await sess.init();
          await sess.login(cfg, email);
          sess.clientNumber = batch.accounts[i].client_number!;
          const keys = await sess.getKeys();
          update(batch, i, {
            public_key: keys.publicKey,
            imagekit_id: keys.imagekitId,
            url_endpoint: `https://ik.imagekit.io/${keys.imagekitId}`,
          });
          await sess.initiateElevation();
          elevated.push({ idx: i, sess });
          console.log(`  ✓ ${email} — OTP sent`);
        } catch (err) {
          fail(batch, i, String(err));
          console.log(`  ✗ ${email} — ${err}`);
          await sess.dispose();
        }
      }
      saveBatch(outputDir, batch);
    }

    // Step 5: Enter OTP codes
    if (elevated.length > 0) {
      console.log(`\n=== STEP 5: Enter OTP codes (${elevated.length} accounts) ===`);
      console.log("Check inbox for 6-digit verification codes.");
      for (const { idx, sess } of elevated) {
        const email = batch.accounts[idx].email;
        const code = await ask(`  OTP for ${email}: `);
        try {
          await sess.verifyElevation(code);
          const fullKeys = await sess.getFullKeys();
          update(batch, idx, { private_key: fullKeys.privateKey });
          advance(batch, idx, "keys_extracted");
          console.log(`  ✓ ${email}`);
        } catch (err) {
          fail(batch, idx, String(err));
          console.log(`  ✗ ${email} — ${err}`);
        } finally {
          await sess.dispose();
        }
      }
      saveBatch(outputDir, batch);
    }
  } finally {
    await context.close();
    await browser.close();
  }

  printFinal(batch);
}

function printSummary(batch: Batch, expected: string) {
  const ok = batch.accounts.filter((a) => a.status === expected).length;
  const failed = batch.accounts.filter((a) => a.status === "failed").length;
  console.log(`  [${ok} ok, ${failed} failed]`);
}

function printFinal(batch: Batch) {
  const done = indicesWithStatus(batch, "keys_extracted").length;
  const failed = indicesWithStatus(batch, "failed").length;
  console.log(`\n=== COMPLETE ===`);
  console.log(`Success: ${done} | Failed: ${failed} | Total: ${batch.accounts.length}`);
  console.log(`State: ${batchPath(outputDir, batch)}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

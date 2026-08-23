import { parseArgs } from "util";
import { mkdtempSync } from "fs";
import { tmpdir } from "os";
import { join } from "path";
import { chromium } from "playwright";
import { loadConfig } from "./config.js";
import { generateEmail } from "./email-gen.js";
import {
  newBatch,
  loadBatch,
  saveBatch,
  setOutputDir,
  batchPath,
  indicesWithStatus,
  type Batch,
} from "./state.js";
import { runAccount, type WorkerCallbacks } from "./worker.js";
import { BROWSER_ARGS, CONTEXT_OPTIONS } from "./stealth.js";
import { configureBuster } from "./buster-setup.js";
import { rootLogger } from "./logger.js";

// @ts-ignore — gmail is JS
import { authorize } from "./gmail/gmail.js";

const { positionals, values } = parseArgs({
  args: process.argv.slice(2),
  options: {
    config: { type: "string", short: "c", default: "config.yaml" },
    output: { type: "string", short: "o", default: "output" },
    concurrency: { type: "string", default: "10" },
  },
  allowPositionals: true,
});

const command = positionals[0] ?? "run";
const cfgPath = values.config!;
const outputDir = values.output!;
const concurrency = parseInt(values.concurrency!, 10) || 3;

setOutputDir(outputDir);

async function main() {
  const cfg = loadConfig(cfgPath);

  // Initialize Gmail API upfront
  rootLogger.info("Connecting to Gmail...");
  await authorize();
  rootLogger.info("Gmail connected.\n");

  let batch: Batch;

  if (command === "run") {
    const emails = Array.from({ length: cfg.batch_size }, () => generateEmail(cfg.email_domain));
    batch = newBatch(emails);
    saveBatch(batch);
    rootLogger.info(`Batch: ${batch.id} (${batch.accounts.length} accounts)`);
    rootLogger.info(`State: ${batchPath(batch)}`);
    rootLogger.info(`Concurrency: ${concurrency}\n`);
  } else if (command === "resume") {
    const batchId = positionals[1];
    if (!batchId) {
      rootLogger.error("Usage: bun src/index.ts resume <batch-id>");
      process.exit(1);
    }
    batch = loadBatch(batchId);
    // Reset FAILED accounts to the best re-entry point based on available data
    let resetCount = 0;
    for (const a of batch.accounts) {
      if (a.status !== "FAILED") continue;
      a.error = undefined;
      if (a.private_key) {
        a.status = "COMPLETED"; // already done
      } else if (a.public_key) {
        a.status = "ONBOARDING_COMPLETED"; // has keys page data, retry OTP reveal
      } else if (a.imagekit_id) {
        a.status = "AUTHENTICATED"; // has imagekit_id from onboarding, retry from there
      } else if (a.client_number) {
        a.status = "EMAIL_VERIFIED"; // registered+verified, retry login
      } else {
        a.status = "CREATED"; // nothing saved, start fresh
      }
      resetCount++;
      rootLogger.info(`  reset [${a.id}] ${a.email}: FAILED → ${a.status}`);
    }
    if (resetCount > 0) {
      saveBatch(batch);
      rootLogger.info(`  (${resetCount} accounts reset)\n`);
    }
    saveBatch(batch);
    rootLogger.info(`Resuming: ${batch.id} (${batch.accounts.length} accounts)`);
    rootLogger.info(`State: ${batchPath(batch)}\n`);
  } else {
    rootLogger.error(`Unknown command: ${command}\nUsage: run | resume <batch-id>`);
    process.exit(1);
  }

  // Find accounts that are not COMPLETED or FAILED
  const pending = batch.accounts
    .map((a, i) => (a.status !== "COMPLETED" ? i : -1))
    .filter((i) => i >= 0);

  if (pending.length === 0) {
    rootLogger.info("All accounts already completed.");
    printSummary(batch);
    process.exit(0);
  }

  rootLogger.info(`Accounts to process: ${pending.length}`);
  rootLogger.info(`Concurrency: ${concurrency}`);
  rootLogger.info("Launching browsers...\n");

  const callbacks: WorkerCallbacks = {};

  async function launchWorkerBrowser(email: string) {
    const dataDir = mkdtempSync(join(tmpdir(), `ik-${email.split("@")[0]}-`));
    const ctx = await chromium.launchPersistentContext(dataDir, {
      headless: false,
      args: BROWSER_ARGS,
      ...CONTEXT_OPTIONS,
    });

    if (cfg.google_stt_api_key) {
      rootLogger.info("Prepare buster");
      await configureBuster(ctx, cfg.google_stt_api_key).catch(() => {});
    }

    return ctx;
  }

  // Launch workers — each picks accounts from the queue sequentially
  let next = 0;

  async function worker(workerId: number) {
    while (next < pending.length) {
      const idx = pending[next++];
      const email = batch.accounts[idx].email;
      rootLogger.info(`Worker ${workerId}: starting ${email}`);

      const ctx = await launchWorkerBrowser(email);
      try {
        await runAccount(ctx, cfg, batch, idx, callbacks);
      } finally {
        await ctx.close().catch(() => {});
      }
    }
  }

  const workerCount = Math.min(concurrency, pending.length);
  await Promise.all(Array.from({ length: workerCount }, (_, i) => worker(i + 1)));

  printSummary(batch);
}

function printSummary(batch: Batch) {
  const completed = indicesWithStatus(batch, "COMPLETED").length;
  const failed = indicesWithStatus(batch, "FAILED").length;
  const other = batch.accounts.length - completed - failed;

  rootLogger.info("=== SUMMARY ===");
  rootLogger.info(
    `Completed: ${completed} | Failed: ${failed} | In-progress: ${other} | Total: ${batch.accounts.length}`,
  );

  if (failed > 0) {
    rootLogger.warn("Failed:");
    for (const a of batch.accounts) {
      if (a.status === "FAILED") rootLogger.warn(`  [${a.id}] ${a.email} — ${a.error}`);
    }
  }
  if (completed > 0) {
    rootLogger.info("Completed:");
    for (const a of batch.accounts) {
      if (a.status === "COMPLETED") rootLogger.info(`  [${a.id}] ${a.email} → ${a.url_endpoint}`);
    }
  }
  rootLogger.info(`State: ${batchPath(batch)}`);
}

main().catch((err) => {
  rootLogger.error(err);
  process.exit(1);
});

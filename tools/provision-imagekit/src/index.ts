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

// @ts-ignore — gmail is JS
import { authorize } from "./gmail/gmail.js";

const { positionals, values } = parseArgs({
  args: process.argv.slice(2),
  options: {
    config: { type: "string", short: "c", default: "config.yaml" },
    output: { type: "string", short: "o", default: "output" },
    concurrency: { type: "string", default: "3" },
    "spawn-delay": { type: "string", default: "10" },
  },
  allowPositionals: true,
});

const command = positionals[0] ?? "run";
const cfgPath = values.config!;
const outputDir = values.output!;
const concurrency = parseInt(values.concurrency!, 10) || 3;
const spawnDelaySec = parseInt(values["spawn-delay"]!, 10) || 10;

setOutputDir(outputDir);

async function main() {
  const cfg = loadConfig(cfgPath);

  // Initialize Gmail API upfront
  console.log("Connecting to Gmail...");
  await authorize();
  console.log("Gmail connected.\n");

  let batch: Batch;

  if (command === "run") {
    const emails = Array.from({ length: cfg.batch_size }, () => generateEmail(cfg.email_domain));
    batch = newBatch(emails);
    saveBatch(batch);
    console.log(`Batch: ${batch.id} (${batch.accounts.length} accounts)`);
    console.log(`State: ${batchPath(batch)}`);
    console.log(`Concurrency: ${concurrency}\n`);
  } else if (command === "resume") {
    const batchId = positionals[1];
    if (!batchId) {
      console.error("Usage: bun src/index.ts resume <batch-id>");
      process.exit(1);
    }
    batch = loadBatch(batchId);
    // Reset FAILED accounts to the best re-entry point based on available data
    let resetCount = 0;
    for (const a of batch.accounts) {
      if (a.status !== "FAILED") continue;
      const prev = a.status;
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
      console.log(`  reset [${a.id}] ${a.email}: FAILED → ${a.status}`);
    }
    if (resetCount > 0) {
      saveBatch(batch);
      console.log(`  (${resetCount} accounts reset)\n`);
    }
    saveBatch(batch);
    console.log(`Resuming: ${batch.id} (${batch.accounts.length} accounts)`);
    console.log(`State: ${batchPath(batch)}\n`);
  } else {
    console.error(`Unknown command: ${command}\nUsage: run | resume <batch-id>`);
    process.exit(1);
  }

  // Find accounts that are not COMPLETED or FAILED
  const pending = batch.accounts
    .map((a, i) => (a.status !== "COMPLETED" ? i : -1))
    .filter((i) => i >= 0);

  if (pending.length === 0) {
    console.log("All accounts already completed.");
    printSummary(batch);
    process.exit(0);
  }

  console.log(`Accounts to process: ${pending.length}`);
  console.log(`Concurrency: ${concurrency} | Spawn delay: ${spawnDelaySec}s`);
  console.log("Launching browsers...\n");

  const callbacks: WorkerCallbacks = {
    log(accountId, msg) {
      console.log(`  ${ts()} [${accountId}] ${msg}`);
    },
  };

  // Each worker gets its own browser instance — full cookie/session isolation
  async function launchWorkerBrowser() {
    const dataDir = mkdtempSync(join(tmpdir(), "imagekit-worker-"));
    const ctx = await chromium.launchPersistentContext(dataDir, {
      headless: false,
      args: BROWSER_ARGS,
      ...CONTEXT_OPTIONS,
    });

    if (cfg.google_stt_api_key) {
      await configureBuster(ctx, cfg.google_stt_api_key).catch(() => {});
    }

    return ctx;
  }

  // Worker pool: run up to `concurrency` accounts simultaneously
  const queue = [...pending];
  const running: Promise<void>[] = [];

  async function spawnNext(): Promise<void> {
    if (queue.length === 0) return;
    const idx = queue.shift()!;
    const ctx = await launchWorkerBrowser();
    const p = runAccount(ctx, cfg, batch, idx, callbacks).then(async () => {
      await ctx.close().catch(() => {});
      const i = running.indexOf(p);
      if (i >= 0) running.splice(i, 1);
      // ponytail: rate-limit between spawns to avoid burst detection
      await sleep(spawnDelaySec * 1000 + Math.random() * 5000);
      spawnNext();
    });
    running.push(p);
  }

  // Seed the pool with staggered starts
  for (let i = 0; i < Math.min(concurrency, queue.length); i++) {
    await spawnNext();
    if (i < Math.min(concurrency, queue.length) - 1) {
      await sleep(spawnDelaySec * 1000 + Math.random() * 3000);
    }
  }

  // Wait for all to finish
  while (running.length > 0) {
    await Promise.race(running);
  }

  printSummary(batch);
}

function printSummary(batch: Batch) {
  const completed = indicesWithStatus(batch, "COMPLETED").length;
  const failed = indicesWithStatus(batch, "FAILED").length;
  const other = batch.accounts.length - completed - failed;
  console.log(`\n=== SUMMARY ===`);
  console.log(
    `Completed: ${completed} | Failed: ${failed} | In-progress: ${other} | Total: ${batch.accounts.length}`,
  );
  if (failed > 0) {
    console.log(`\nFailed:`);
    for (const a of batch.accounts) {
      if (a.status === "FAILED") console.log(`  [${a.id}] ${a.email} — ${a.error}`);
    }
  }
  if (completed > 0) {
    console.log(`\nCompleted:`);
    for (const a of batch.accounts) {
      if (a.status === "COMPLETED") console.log(`  [${a.id}] ${a.email} → ${a.url_endpoint}`);
    }
  }
  console.log(`\nState: ${batchPath(batch)}`);
}

function ts(): string {
  return new Date().toISOString().slice(11, 19);
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

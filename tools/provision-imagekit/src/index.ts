import { parseArgs } from "util";
import { createInterface } from "readline";
import { chromium } from "playwright";
import { loadConfig } from "./config.js";
import { generateEmail } from "./email-gen.js";
import {
  newBatch, loadBatch, saveBatch, setOutputDir, batchPath,
  indicesWithStatus, type Batch,
} from "./state.js";
import { runAccount, type WorkerCallbacks } from "./worker.js";

// @ts-ignore — gmail is JS
import { authorize } from "./gmail/gmail.js";

const { positionals, values } = parseArgs({
  args: process.argv.slice(2),
  options: {
    config: { type: "string", short: "c", default: "config.yaml" },
    output: { type: "string", short: "o", default: "output" },
    concurrency: { type: "string", default: "3" },
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
  console.log("Launching browser...\n");

  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();

  // Readline for human intervention prompts
  const rl = createInterface({ input: process.stdin, output: process.stdout });

  const callbacks: WorkerCallbacks = {
    async onHumanRequired(accountId, email, reason) {
      console.log(`\n  ⚠️  [${accountId}] ${email} — HUMAN REQUIRED: ${reason}`);
      console.log(`      Solve it in the browser, then press Enter here.`);
      await new Promise<void>((resolve) => rl.question("      Press Enter when done... ", () => resolve()));
    },
    log(accountId, msg) {
      console.log(`  [${accountId}] ${msg}`);
    },
  };

  // Worker pool: run up to `concurrency` accounts simultaneously
  const queue = [...pending];
  const running: Promise<void>[] = [];

  function spawnNext(): void {
    if (queue.length === 0) return;
    const idx = queue.shift()!;
    const p = runAccount(context, cfg, batch, idx, callbacks).then(() => {
      // Remove from running, spawn next
      const i = running.indexOf(p);
      if (i >= 0) running.splice(i, 1);
      spawnNext();
    });
    running.push(p);
  }

  // Seed the pool
  for (let i = 0; i < Math.min(concurrency, queue.length); i++) {
    spawnNext();
  }

  // Wait for all to finish
  while (running.length > 0) {
    await Promise.race(running);
  }

  rl.close();
  await context.close();
  await browser.close();

  printSummary(batch);
}

function printSummary(batch: Batch) {
  const completed = indicesWithStatus(batch, "COMPLETED").length;
  const failed = indicesWithStatus(batch, "FAILED").length;
  const other = batch.accounts.length - completed - failed;
  console.log(`\n=== SUMMARY ===`);
  console.log(`Completed: ${completed} | Failed: ${failed} | In-progress: ${other} | Total: ${batch.accounts.length}`);
  console.log(`State: ${batchPath(batch)}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

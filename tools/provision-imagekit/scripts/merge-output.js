/**
 * Merge all batch output files into:
 *   output/00_completed.json  — accounts with status COMPLETED (have all keys)
 *   output/00_retryable.json  — accounts that can be resumed (registered but not completed)
 *
 * Unique by email. Latest entry wins on duplicates.
 * Run: bun scripts/merge-output.js
 */

import { readdirSync, readFileSync, writeFileSync } from "fs";
import { join } from "path";

const OUTPUT_DIR = "output";
const COMPLETED_FILE = join(OUTPUT_DIR, "00_completed.json");
const RETRYABLE_FILE = join(OUTPUT_DIR, "00_retryable.json");

// States that mean "done"
const DONE_STATES = new Set(["COMPLETED"]);

// States that can be retried (have some progress)
const RETRYABLE_STATES = new Set([
  "REGISTERED",
  "EMAIL_VERIFICATION_PENDING",
  "EMAIL_VERIFIED",
  "LOGIN",
  "LOGIN_HUMAN_REQUIRED",
  "AUTHENTICATED",
  "ONBOARDING_REQUIRED",
  "ONBOARDING_COMPLETED",
  "CONFIG_RETRIEVAL",
  "PRIVATE_KEY_OTP_PENDING",
  "PRIVATE_KEY_REVEALED",
  // Also include FAILED that have useful data
  "FAILED",
]);

function main() {
  const files = readdirSync(OUTPUT_DIR)
    .filter((f) => f.startsWith("batch_") && f.endsWith(".json"))
    .sort();

  /** @type {Map<string, object>} */
  const completedMap = new Map();
  /** @type {Map<string, object>} */
  const retryableMap = new Map();

  let totalFiles = 0;
  let totalAccounts = 0;

  for (const file of files) {
    let data;
    try {
      data = JSON.parse(readFileSync(join(OUTPUT_DIR, file), "utf-8"));
    } catch {
      continue;
    }

    totalFiles++;
    const accounts = data.accounts || [];

    for (const account of accounts) {
      totalAccounts++;
      const email = account.email;
      if (!email) continue;

      if (DONE_STATES.has(account.status)) {
        // Completed — always takes priority
        completedMap.set(email, account);
        retryableMap.delete(email); // remove from retryable if previously added
      } else if (!completedMap.has(email)) {
        // Not completed — check if retryable (has any useful data)
        const hasData = account.client_number || account.imagekit_id || account.public_key;
        const isRetryableState = RETRYABLE_STATES.has(account.status);

        if (hasData || isRetryableState) {
          // Keep the entry with the most progress
          const existing = retryableMap.get(email);
          if (!existing || hasMoreProgress(account, existing)) {
            retryableMap.set(email, account);
          }
        }
      }
    }
  }

  // Determine best resume state for retryable accounts
  const retryable = [...retryableMap.values()].map((a) => ({
    ...a,
    resume_from: determineResumeState(a),
  }));

  const completed = [...completedMap.values()];

  writeFileSync(COMPLETED_FILE, JSON.stringify({ accounts: completed }, null, 2));
  writeFileSync(RETRYABLE_FILE, JSON.stringify({ accounts: retryable }, null, 2));

  console.log(`Scanned ${totalFiles} batch files, ${totalAccounts} total account entries`);
  console.log(`\nCompleted: ${completed.length} → ${COMPLETED_FILE}`);
  console.log(`Retryable: ${retryable.length} → ${RETRYABLE_FILE}`);

  if (completed.length > 0) {
    console.log("\n--- Completed ---");
    for (const a of completed) {
      console.log(`  ${a.email} | ${a.imagekit_id} | ${a.url_endpoint}`);
    }
  }

  if (retryable.length > 0) {
    console.log("\n--- Retryable ---");
    for (const a of retryable) {
      console.log(`  ${a.email} | status=${a.status} | resume=${a.resume_from} | ik_id=${a.imagekit_id || "-"}`);
    }
  }
}

function hasMoreProgress(a, b) {
  const score = (acc) => {
    if (acc.private_key) return 6;
    if (acc.public_key) return 5;
    if (acc.imagekit_id) return 4;
    if (acc.client_number) return 3;
    return 1;
  };
  return score(a) >= score(b);
}

function determineResumeState(account) {
  if (account.private_key) return "COMPLETED";
  if (account.public_key) return "ONBOARDING_COMPLETED";
  if (account.imagekit_id) return "AUTHENTICATED";
  if (account.client_number) return "EMAIL_VERIFIED";
  return "CREATED";
}

main();

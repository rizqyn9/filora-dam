import { authorize, searchMessages } from "../src/gmail/gmail.js";
import { readdirSync, readFileSync } from "fs";
import { join } from "path";

async function main() {
  console.log("Connecting to Gmail...");
  await authorize();

  // Collect all tracked emails from batch files
  const tracked = new Set();
  const outputDir = "output";
  const files = readdirSync(outputDir).filter((f) => f.endsWith(".json"));
  for (const f of files) {
    try {
      const data = JSON.parse(readFileSync(join(outputDir, f), "utf-8"));
      for (const a of data.accounts) {
        if (a.status === "COMPLETED" || a.status === "ONBOARDING_COMPLETED") {
          tracked.add(a.email);
        }
      }
    } catch {
      // skip corrupted files
    }
  }
  console.log(`Already tracked (completed/recovery): ${tracked.size}\n`);

  // Search Gmail for all ImageKit verification emails to our domain
  const result = await searchMessages({
    query: "from:imagekit.io to:dam.ke-gap-bocil.my.id subject:Account Verification",
    maxResults: 50,
  });

  console.log(`Gmail verification emails found: ${result.messages.length}\n`);

  const untracked = [];
  const seen = new Set();

  for (const msg of result.messages) {
    const to = msg.to || "";
    const match = to.match(/([\w.-]+@dam\.ke-gap-bocil\.my\.id)/);
    const email = match ? match[1] : to.trim();
    if (!email || tracked.has(email) || seen.has(email)) continue;
    seen.add(email);
    untracked.push({ email, date: msg.date, subject: msg.subject });
  }

  console.log("=== UNTRACKED REGISTERED ACCOUNTS ===");
  for (const u of untracked) {
    console.log(`  ${u.email} | ${u.date}`);
  }
  console.log(`\nTotal: ${untracked.length}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

import { readFileSync, writeFileSync, mkdirSync } from "fs";
import { join } from "path";

export type Status =
  | "pending"
  | "registered"
  | "verified"
  | "onboarded"
  | "keys_extracted"
  | "failed";

export interface Account {
  email: string;
  status: Status;
  error?: string;
  client_number?: string;
  imagekit_id?: string;
  public_key?: string;
  private_key?: string;
  url_endpoint?: string;
  created_at: string;
  updated_at: string;
}

export interface Batch {
  id: string;
  accounts: Account[];
}

export function newBatch(emails: string[]): Batch {
  const now = new Date().toISOString();
  const ts = now.replace(/[-:T.]/g, "").slice(0, 14);
  return {
    id: `batch_${ts}`,
    accounts: emails.map((email) => ({
      email,
      status: "pending",
      created_at: now,
      updated_at: now,
    })),
  };
}

export function advance(batch: Batch, idx: number, status: Status): void {
  batch.accounts[idx].status = status;
  batch.accounts[idx].updated_at = new Date().toISOString();
}

export function fail(batch: Batch, idx: number, err: string): void {
  batch.accounts[idx].status = "failed";
  batch.accounts[idx].error = err;
  batch.accounts[idx].updated_at = new Date().toISOString();
}

export function update(batch: Batch, idx: number, fields: Partial<Account>): void {
  Object.assign(batch.accounts[idx], fields);
  batch.accounts[idx].updated_at = new Date().toISOString();
}

export function indicesWithStatus(batch: Batch, status: Status): number[] {
  return batch.accounts.map((a, i) => (a.status === status ? i : -1)).filter((i) => i >= 0);
}

export function saveBatch(outputDir: string, batch: Batch): void {
  mkdirSync(outputDir, { recursive: true });
  const path = join(outputDir, `${batch.id}.json`);
  writeFileSync(path, JSON.stringify(batch, null, 2), "utf-8");
}

export function loadBatch(outputDir: string, batchId: string): Batch {
  const path = join(outputDir, `${batchId}.json`);
  return JSON.parse(readFileSync(path, "utf-8"));
}

export function batchPath(outputDir: string, batch: Batch): string {
  return join(outputDir, `${batch.id}.json`);
}

import { readFileSync, writeFileSync, mkdirSync, existsSync } from "fs";
import { join } from "path";

export type Status =
  | "CREATED"
  | "REGISTRATION"
  | "REGISTRATION_HUMAN_REQUIRED"
  | "EMAIL_VERIFICATION_PENDING"
  | "EMAIL_VERIFIED"
  | "LOGIN"
  | "LOGIN_HUMAN_REQUIRED"
  | "AUTHENTICATED"
  | "ONBOARDING_REQUIRED"
  | "ONBOARDING_COMPLETED"
  | "CONFIG_RETRIEVAL"
  | "PRIVATE_KEY_OTP_PENDING"
  | "PRIVATE_KEY_REVEALED"
  | "COMPLETED"
  | "FAILED";

export interface Account {
  id: string;
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

// --- Mutex for serialized writes ---

let writeChain: Promise<void> = Promise.resolve();

function serializedWrite(fn: () => void): Promise<void> {
  writeChain = writeChain.then(fn, fn);
  return writeChain;
}

// --- Batch operations ---

export function newBatch(emails: string[]): Batch {
  const now = new Date().toISOString();
  const ts = now.replace(/[-:T.Z]/g, "").slice(0, 14);
  return {
    id: `batch_${ts}`,
    accounts: emails.map((email, i) => ({
      id: `A${String(i + 1).padStart(3, "0")}`,
      email,
      status: "CREATED",
      created_at: now,
      updated_at: now,
    })),
  };
}

export function transition(batch: Batch, idx: number, status: Status): Promise<void> {
  batch.accounts[idx].status = status;
  batch.accounts[idx].updated_at = new Date().toISOString();
  // Error is cleared on forward transition
  if (status !== "FAILED") {
    batch.accounts[idx].error = undefined;
  }
  return serializedWrite(() => writeBatch(batch));
}

export function fail(batch: Batch, idx: number, error: string): Promise<void> {
  batch.accounts[idx].status = "FAILED";
  batch.accounts[idx].error = error;
  batch.accounts[idx].updated_at = new Date().toISOString();
  return serializedWrite(() => writeBatch(batch));
}

export function update(batch: Batch, idx: number, fields: Partial<Account>): Promise<void> {
  Object.assign(batch.accounts[idx], fields);
  batch.accounts[idx].updated_at = new Date().toISOString();
  return serializedWrite(() => writeBatch(batch));
}

export function indicesWithStatus(batch: Batch, ...statuses: Status[]): number[] {
  return batch.accounts
    .map((a, i) => (statuses.includes(a.status) ? i : -1))
    .filter((i) => i >= 0);
}

// --- Persistence ---

let _outputDir = "output";

export function setOutputDir(dir: string): void {
  _outputDir = dir;
}

export function batchPath(batch: Batch): string {
  return join(_outputDir, `${batch.id}.json`);
}

function writeBatch(batch: Batch): void {
  mkdirSync(_outputDir, { recursive: true });
  writeFileSync(join(_outputDir, `${batch.id}.json`), JSON.stringify(batch, null, 2), "utf-8");
}

export function saveBatch(batch: Batch): void {
  writeBatch(batch);
}

export function loadBatch(batchId: string): Batch {
  const path = join(_outputDir, `${batchId}.json`);
  if (!existsSync(path)) throw new Error(`Batch not found: ${path}`);
  return JSON.parse(readFileSync(path, "utf-8"));
}

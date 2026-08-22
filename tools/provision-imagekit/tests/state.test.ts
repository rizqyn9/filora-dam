import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { rmSync, existsSync } from "fs";
import {
  newBatch,
  advance,
  fail,
  update,
  indicesWithStatus,
  saveBatch,
  loadBatch,
} from "../src/state.js";

const TEST_DIR = "output/_test";

beforeEach(() => {
  rmSync(TEST_DIR, { recursive: true, force: true });
});
afterEach(() => {
  rmSync(TEST_DIR, { recursive: true, force: true });
});

describe("state", () => {
  test("newBatch creates accounts with pending status", () => {
    const batch = newBatch(["a@x.com", "b@x.com"]);
    expect(batch.id).toMatch(/^batch_\d/);
    expect(batch.accounts).toHaveLength(2);
    expect(batch.accounts[0].status).toBe("pending");
    expect(batch.accounts[0].email).toBe("a@x.com");
  });

  test("advance updates status", () => {
    const batch = newBatch(["a@x.com"]);
    advance(batch, 0, "registered");
    expect(batch.accounts[0].status).toBe("registered");
  });

  test("fail sets status and error", () => {
    const batch = newBatch(["a@x.com"]);
    fail(batch, 0, "something broke");
    expect(batch.accounts[0].status).toBe("failed");
    expect(batch.accounts[0].error).toBe("something broke");
  });

  test("update merges fields", () => {
    const batch = newBatch(["a@x.com"]);
    update(batch, 0, { client_number: "ABC123", public_key: "pub_x" });
    expect(batch.accounts[0].client_number).toBe("ABC123");
    expect(batch.accounts[0].public_key).toBe("pub_x");
  });

  test("indicesWithStatus filters correctly", () => {
    const batch = newBatch(["a@x.com", "b@x.com", "c@x.com"]);
    advance(batch, 0, "registered");
    advance(batch, 2, "registered");
    expect(indicesWithStatus(batch, "registered")).toEqual([0, 2]);
    expect(indicesWithStatus(batch, "pending")).toEqual([1]);
  });

  test("save and load roundtrip", () => {
    const batch = newBatch(["a@x.com"]);
    advance(batch, 0, "registered");
    update(batch, 0, { client_number: "XYZ" });
    saveBatch(TEST_DIR, batch);
    expect(existsSync(`${TEST_DIR}/${batch.id}.json`)).toBe(true);

    const loaded = loadBatch(TEST_DIR, batch.id);
    expect(loaded.id).toBe(batch.id);
    expect(loaded.accounts[0].client_number).toBe("XYZ");
    expect(loaded.accounts[0].status).toBe("registered");
  });
});

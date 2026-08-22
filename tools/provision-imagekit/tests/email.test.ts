import { describe, test, expect } from "bun:test";
import { generateEmail } from "../src/email.js";

describe("generateEmail", () => {
  test("produces valid email format", () => {
    const email = generateEmail("img.example.com");
    expect(email).toMatch(/^[a-z]+[a-z]+\d{3}@img\.example\.com$/);
  });

  test("uses the provided domain", () => {
    const email = generateEmail("dam.ke-gap-bocil.my.id");
    expect(email.endsWith("@dam.ke-gap-bocil.my.id")).toBe(true);
  });

  test("generates different emails on repeated calls", () => {
    const emails = new Set(
      Array.from({ length: 20 }, () => generateEmail("x.com"))
    );
    // With 32*32*900 = ~921k combinations, 20 should be unique
    expect(emails.size).toBeGreaterThan(15);
  });

  test("local part is lowercase alpha + digits only", () => {
    for (let i = 0; i < 50; i++) {
      const local = generateEmail("d.com").split("@")[0];
      expect(local).toMatch(/^[a-z]+\d{3}$/);
    }
  });
});

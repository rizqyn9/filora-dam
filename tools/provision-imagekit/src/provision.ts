import { type BrowserContext, type Page } from "playwright";
import type { Config } from "./config.js";
import { regionCode } from "./config.js";

export interface DevKeys {
  publicKey: string;
  privateKey: string;
  imagekitId: string;
}

/**
 * All actions via real browser interactions — no fetch, no raw HTTP.
 */
export class Session {
  public clientNumber = "";
  private page!: Page;

  constructor(private ctx: BrowserContext) {}

  async init(): Promise<void> {
    this.page = await this.ctx.newPage();
  }

  async dispose(): Promise<void> {
    await this.page.close().catch(() => {});
  }

  async login(cfg: Config, email: string): Promise<void> {
    await this.page.goto("https://imagekit.io/login/", { waitUntil: "commit" });
    await this.page.waitForSelector('input[placeholder="Enter your work email"]', {
      state: "visible",
      timeout: 60_000,
    });

    await this.page.locator('input[placeholder="Enter your work email"]').fill(email);
    await this.page.locator('input[placeholder="Enter a password"]').fill(cfg.password);
    await this.page.getByRole("button", { name: "Submit" }).click();

    // Wait for navigation away from login
    await this.page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 60_000 });
  }

  async completeOnboarding(cfg: Config): Promise<void> {
    const url = this.page.url();
    if (!url.includes("get-started")) {
      await this.page.goto("https://imagekit.io/dashboard/get-started", { waitUntil: "commit" });
    }

    // 1/3: Role + objective
    await this.page.waitForSelector("select", { state: "visible", timeout: 30_000 });
    await this.page.locator("select").selectOption({ index: 1 });
    await this.page.getByText("Real-time image optimization").click();
    await this.page.getByRole("button", { name: "Next" }).click();

    // 2/3: Region
    await this.page.waitForSelector("select", { state: "visible", timeout: 15_000 });
    await this.page.locator("select").selectOption(regionCode(cfg.region));
    await this.page.getByRole("button", { name: "Next" }).click();

    // 3/3: How did you hear
    await this.page.waitForTimeout(2000);
    await this.page.getByText("Google / Bing").click();
    await this.page.getByRole("button", { name: "Submit" }).click();

    // Wait for dashboard
    await this.page.waitForURL((url) => !url.pathname.includes("get-started"), { timeout: 30_000 });
  }

  async getKeys(): Promise<DevKeys> {
    await this.page.goto("https://imagekit.io/dashboard/developer/api-keys", {
      waitUntil: "commit",
    });
    await this.page.waitForSelector('input[value^="public_"]', {
      state: "visible",
      timeout: 30_000,
    });

    const publicKey = await this.page.locator('input[value^="public_"]').inputValue();
    const endpointText =
      (await this.page.locator("text=https://ik.imagekit.io/").textContent()) ?? "";
    const match = endpointText.match(/https:\/\/ik\.imagekit\.io\/\S+/);
    const imagekitId = match ? match[0].split("/").pop()! : "";

    return { publicKey, privateKey: "", imagekitId };
  }

  async initiateElevation(): Promise<void> {
    // Click the eye icon next to private key
    await this.page.locator("td:nth-child(2) .chakra-icon").first().click();
    // Wait for OTP modal
    await this.page.waitForSelector('[aria-modal="true"]', { state: "visible", timeout: 10_000 });
  }

  async verifyElevation(code: string): Promise<void> {
    // Type OTP digits into pin inputs
    const digits = code.split("");
    const inputs = this.page.locator('[aria-modal="true"] input');
    for (let i = 0; i < digits.length; i++) {
      await inputs.nth(i).fill(digits[i]);
    }
    // Click Verify
    await this.page.getByRole("button", { name: "Verify" }).click();
    await this.page.waitForTimeout(2000);
  }

  async getFullKeys(): Promise<DevKeys> {
    // After elevation, private key should be visible
    await this.page.waitForTimeout(1000);
    const publicKey = await this.page.locator('input[value^="public_"]').inputValue();
    const privateKey = await this.page.locator('input[value^="private_"]').inputValue();

    const endpointText =
      (await this.page.locator("text=https://ik.imagekit.io/").textContent()) ?? "";
    const match = endpointText.match(/https:\/\/ik\.imagekit\.io\/\S+/);
    const imagekitId = match ? match[0].split("/").pop()! : "";

    return { publicKey, privateKey, imagekitId };
  }
}

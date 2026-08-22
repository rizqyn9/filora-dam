import { readFileSync } from "fs";
import { parse } from "yaml";

export interface Config {
  email_domain: string;
  password: string;
  name: string;
  batch_size: number;
  region: string;
}

export function loadConfig(path: string): Config {
  const raw = readFileSync(path, "utf-8");
  const cfg = parse(raw) as Config;

  if (!cfg.email_domain) throw new Error("email_domain required");
  if (!cfg.password || cfg.password.length < 8)
    throw new Error("password required (min 8 chars)");
  if (!cfg.name) throw new Error("name required");
  cfg.batch_size = cfg.batch_size || 1;
  cfg.region = cfg.region || "Singapore";

  return cfg;
}

export function regionCode(region: string): string {
  if (region.includes("Singapore")) return "sgp";
  if (region.includes("Mumbai")) return "ind";
  if (region.includes("Frankfurt")) return "eu";
  if (region.includes("Virginia")) return "us-east-1";
  if (region.includes("California")) return "us-west-1";
  if (region.includes("Sydney")) return "au";
  if (region.includes("Montreal")) return "ca";
  return "sgp";
}

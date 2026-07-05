import { z } from "zod";

/**
 * Validated environment variables (fail fast at startup).
 * All web env vars must be prefixed with VITE_ to be exposed by Vite.
 */
const envSchema = z.object({
  VITE_API_BASE_URL: z.string().min(1).default("/api/v1"),
  // Clerk publishable key (pk_test_… / pk_live_…). Owns web sign-in.
  VITE_CLERK_PUBLISHABLE_KEY: z.string().min(1),
});

const parsed = envSchema.safeParse(import.meta.env);

if (!parsed.success) {
  console.error("Invalid environment variables:", z.treeifyError(parsed.error));
  throw new Error("Invalid environment variables");
}

export const env = parsed.data;

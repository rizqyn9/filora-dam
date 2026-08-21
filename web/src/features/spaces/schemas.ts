import { z } from "zod";

export const SpaceSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  owner_id: z.number(),
  storage_quota_bytes: z.number(),
  storage_used_bytes: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type Space = z.infer<typeof SpaceSchema>;

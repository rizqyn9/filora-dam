import { z } from "zod";

/** Mirrors apps/api account.User. */
export const userSchema = z.object({
  id: z.number().int(),
  clerk_user_id: z.string(),
  email: z.email(),
  name: z.string(),
  avatar_url: z.string().nullish(),
  is_active: z.boolean(),
  last_seen_at: z.iso.datetime().nullish(),
  created_at: z.iso.datetime(),
  updated_at: z.iso.datetime(),
});

export type User = z.infer<typeof userSchema>;

/** Payload for PATCH /me. */
export const updateProfileSchema = z.object({
  name: z.string().min(1).max(255),
  avatar_url: z.url().nullish(),
});

export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;

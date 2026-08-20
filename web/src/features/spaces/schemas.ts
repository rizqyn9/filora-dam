import { z } from "zod";

/** Mirrors api/ space.Space. */
export const spaceSchema = z.object({
  id: z.number().int(),
  owner_id: z.number().int(),
  name: z.string(),
  description: z.string().nullish(),
  is_default: z.boolean(),
  storage_quota: z.number().int(),
  storage_used: z.number().int(),
  created_at: z.iso.datetime(),
  updated_at: z.iso.datetime(),
});

export type Space = z.infer<typeof spaceSchema>;

export const spaceListSchema = z.array(spaceSchema);

/** POST /spaces */
export const createSpaceSchema = z.object({
  name: z.string().min(1).max(255),
  description: z.string().max(1000).nullish(),
});

export type CreateSpaceInput = z.infer<typeof createSpaceSchema>;

/** PATCH /spaces/:id */
export const updateSpaceSchema = createSpaceSchema;

export type UpdateSpaceInput = z.infer<typeof updateSpaceSchema>;

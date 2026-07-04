import { z } from "zod";

/** Mirrors apps/api gallery.Gallery. */
export const gallerySchema = z.object({
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

export type Gallery = z.infer<typeof gallerySchema>;

export const gallListSchema = z.array(gallerySchema);

/** POST /galleries */
export const createGallerySchema = z.object({
  name: z.string().min(1).max(255),
  description: z.string().max(1000).nullish(),
});

export type CreateGalleryInput = z.infer<typeof createGallerySchema>;

/** PATCH /galleries/:id */
export const updateGallerySchema = createGallerySchema;

export type UpdateGalleryInput = z.infer<typeof updateGallerySchema>;

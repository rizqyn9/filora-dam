import { z } from "zod";

/** Mirrors api/ asset.Asset. */
export const assetSchema = z.object({
  id: z.string().uuid(),
  space_id: z.number().int(),
  folder_id: z.number().int().nullish(),
  uploaded_by: z.number().int().nullish(),
  name: z.string(),
  type: z.string(),
  mime_type: z.string(),
  size: z.number().int(),
  hash: z.string(),
  metadata: z.unknown().nullish(),
  deleted_at: z.iso.datetime().nullish(),
  created_at: z.iso.datetime(),
  updated_at: z.iso.datetime(),
});

export type Asset = z.infer<typeof assetSchema>;

export const assetListResultSchema = z.object({
  assets: z.array(assetSchema),
  total: z.number().int(),
  limit: z.number().int(),
  offset: z.number().int(),
});

export type AssetListResult = z.infer<typeof assetListResultSchema>;

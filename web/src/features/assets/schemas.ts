import { z } from "zod";

export const AssetSchema = z.object({
  id: z.string().uuid(),
  original_filename: z.string(),
  name: z.string(),
  mime_type: z.string(),
  size_bytes: z.number(),
  checksum_sha256: z.string(),
  uploaded_by: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type Asset = z.infer<typeof AssetSchema>;

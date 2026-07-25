import { z } from "zod";

/** Mirrors apps/api folder.Folder. */
export const folderSchema = z.object({
  id: z.number().int(),
  space_id: z.number().int(),
  parent_id: z.number().int().nullish(),
  owner_id: z.number().int(),
  name: z.string(),
  path: z.string(),
  created_at: z.iso.datetime(),
  updated_at: z.iso.datetime(),
});

export type Folder = z.infer<typeof folderSchema>;

export const folderListSchema = z.array(folderSchema);

export const breadcrumbItemSchema = z.object({
  id: z.number().int(),
  name: z.string(),
});

export type BreadcrumbItem = z.infer<typeof breadcrumbItemSchema>;

export const breadcrumbListSchema = z.array(breadcrumbItemSchema);

/** POST /spaces/:spaceId/folders */
export const createFolderSchema = z.object({
  name: z.string().min(1).max(255),
  parent_id: z.number().int().nullish(),
});

export type CreateFolderInput = z.infer<typeof createFolderSchema>;

/** PATCH /folders/:id */
export const updateFolderSchema = z.object({
  name: z.string().min(1).max(255),
});

export type UpdateFolderInput = z.infer<typeof updateFolderSchema>;

/** POST /folders/:id/move */
export const moveFolderSchema = z.object({
  parent_id: z.number().int().nullish(),
});

export type MoveFolderInput = z.infer<typeof moveFolderSchema>;

import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";

import { FolderSchema, type Folder } from "./schemas";

export function useFolders(spaceId: string | undefined) {
  return useQuery<Folder[]>({
    queryKey: ["folders", spaceId],
    queryFn: async () => {
      const raw = await api<unknown[]>(`/folders?space_id=${spaceId}`);
      return raw.map((f) => FolderSchema.parse(f));
    },
    enabled: !!spaceId,
  });
}

export interface BreadcrumbItem {
  id: string;
  name: string;
}

export function useBreadcrumbs(folderId: string | undefined) {
  return useQuery<BreadcrumbItem[]>({
    queryKey: ["breadcrumbs", folderId],
    queryFn: async () => {
      const raw = await api<unknown[]>(`/folders/${folderId}/breadcrumbs`);
      return raw.map((b) =>
        FolderSchema.pick({ id: true, name: true }).parse(b),
      );
    },
    enabled: !!folderId,
  });
}

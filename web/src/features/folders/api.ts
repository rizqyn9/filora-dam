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

import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";

import { AssetSchema, type Asset } from "./schemas";

interface UseAssetsParams {
  spaceId: string;
  folderId?: string;
  limit?: number;
  offset?: number;
}

export function useAssets({ spaceId, folderId, limit = 50, offset = 0 }: UseAssetsParams) {
  return useQuery<Asset[]>({
    queryKey: ["assets", spaceId, folderId ?? "root", limit, offset],
    queryFn: async () => {
      const params = new URLSearchParams({
        space_id: spaceId,
        limit: String(limit),
        offset: String(offset),
      });
      if (folderId) params.set("folder_id", folderId);

      const raw = await api<unknown[]>(`/assets?${params}`);
      return raw.map((a) => AssetSchema.parse(a));
    },
    enabled: !!spaceId,
  });
}

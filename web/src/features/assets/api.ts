import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { assetListResultSchema, type Asset, type AssetListResult } from "@/features/assets/schemas";
import { api } from "@/lib/api-client";

export const assetKeys = {
  all: ["assets"] as const,
  lists: () => [...assetKeys.all, "list"] as const,
  bySpace: (spaceId: number, folderId?: number | null) =>
    [...assetKeys.lists(), spaceId, folderId ?? "root"] as const,
  allInSpace: (spaceId: number) =>
    [...assetKeys.lists(), spaceId, "all"] as const,
  byType: (spaceId: number, type: string) =>
    [...assetKeys.lists(), spaceId, "type", type] as const,
  search: (spaceId: number, query: string) =>
    [...assetKeys.lists(), spaceId, "search", query] as const,
  trash: (spaceId: number) =>
    [...assetKeys.lists(), spaceId, "trash"] as const,
  details: () => [...assetKeys.all, "detail"] as const,
  detail: (id: string) => [...assetKeys.details(), id] as const,
};

export const assetsQueryOptions = (
  spaceId: number,
  folderId?: number | null,
  page = { limit: 50, offset: 0 },
) =>
  queryOptions({
    queryKey: [...assetKeys.bySpace(spaceId, folderId), page] as const,
    queryFn: async () => {
      const params: Record<string, string | number> = {
        limit: page.limit,
        offset: page.offset,
      };
      if (folderId != null) params.folder_id = folderId;
      return assetListResultSchema.parse(
        await api.get<AssetListResult>(`/spaces/${spaceId}/assets`, { params }),
      );
    },
  });

export const assetsByTypeQueryOptions = (
  spaceId: number,
  type: string,
  page = { limit: 50, offset: 0 },
) =>
  queryOptions({
    queryKey: [...assetKeys.byType(spaceId, type), page] as const,
    queryFn: async () =>
      assetListResultSchema.parse(
        await api.get<AssetListResult>(`/spaces/${spaceId}/assets/filter/${type}`, {
          params: { limit: page.limit, offset: page.offset },
        }),
      ),
  });

export function useAssets(spaceId: number, folderId?: number | null) {
  return useQuery(assetsQueryOptions(spaceId, folderId));
}

export function useAssetsByType(spaceId: number, type: string) {
  return useQuery(assetsByTypeQueryOptions(spaceId, type));
}

export function useMoveAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, folderId }: { id: string; folderId: number | null }) =>
      api.post<Asset>(`/assets/${id}/move`, { folder_id: folderId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: assetKeys.lists() });
    },
  });
}

export function useDeleteAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/assets/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: assetKeys.lists() });
    },
  });
}

export function useRestoreAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post<void>(`/assets/${id}/restore`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: assetKeys.lists() });
    },
  });
}

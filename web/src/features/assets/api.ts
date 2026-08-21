import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";

import { AssetSchema, type Asset } from "./schemas";

interface UseAssetsParams {
  spaceId: string;
  folderId?: string;
  limit?: number;
  offset?: number;
}

export function useAssets({
  spaceId,
  folderId,
  limit = 50,
  offset = 0,
}: UseAssetsParams) {
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

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { useAuthStore } from "@/features/auth/auth-store";
import { ApiRequestError } from "@/lib/api";

interface UploadFileInput {
  spaceId: string;
  folderId?: string;
  file: File;
}

async function uploadFile({
  spaceId,
  folderId,
  file,
}: UploadFileInput): Promise<Asset> {
  const token = useAuthStore.getState().token;
  const form = new FormData();
  form.append("space_id", spaceId);
  if (folderId) form.append("folder_id", folderId);
  form.append("file", file);

  const res = await fetch("/api/v1/assets/upload", {
    method: "POST",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  });

  const body = await res.json();
  if (!res.ok || !body.success) {
    throw new ApiRequestError(res.status, body.error);
  }
  return AssetSchema.parse(body.data);
}

export function useUpload(spaceId: string, folderId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (file: File) => uploadFile({ spaceId, folderId, file }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["assets", spaceId] });
    },
  });
}

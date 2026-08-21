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

import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useCreateFolder(spaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { name: string; parentId?: string }) =>
      api<Folder>("/folders", {
        method: "POST",
        body: JSON.stringify({
          space_id: spaceId,
          parent_id: input.parentId ?? null,
          name: input.name,
        }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders", spaceId] });
    },
  });
}

export function useRenameFolder(spaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ folderId, name }: { folderId: string; name: string }) =>
      api(`/folders/${folderId}/rename`, {
        method: "PATCH",
        body: JSON.stringify({ name }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders", spaceId] });
    },
  });
}

export function useMoveFolder(spaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      folderId,
      parentId,
    }: {
      folderId: string;
      parentId: string | null;
    }) =>
      api(`/folders/${folderId}/move`, {
        method: "PATCH",
        body: JSON.stringify({ parent_id: parentId }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders", spaceId] });
    },
  });
}

export function useDeleteFolder(spaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (folderId: string) =>
      api(`/folders/${folderId}`, { method: "DELETE" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders", spaceId] });
    },
  });
}

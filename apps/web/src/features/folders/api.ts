import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import {
  breadcrumbListSchema,
  folderListSchema,
  folderSchema,
  type BreadcrumbItem,
  type CreateFolderInput,
  type Folder,
  type MoveFolderInput,
  type UpdateFolderInput,
} from "@/features/folders/schemas";
import { api } from "@/lib/api-client";

export const folderKeys = {
  all: ["folders"] as const,
  lists: () => [...folderKeys.all, "list"] as const,
  bySpace: (spaceId: number, parentId?: number | null) =>
    [...folderKeys.lists(), spaceId, parentId ?? "root"] as const,
  details: () => [...folderKeys.all, "detail"] as const,
  detail: (id: number) => [...folderKeys.details(), id] as const,
  breadcrumb: (id: number) => [...folderKeys.all, "breadcrumb", id] as const,
  children: (id: number) => [...folderKeys.all, "children", id] as const,
};

export const foldersQueryOptions = (
  spaceId: number,
  parentId?: number | null,
) =>
  queryOptions({
    queryKey: folderKeys.bySpace(spaceId, parentId),
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (parentId != null) params.parent_id = String(parentId);
      return folderListSchema.parse(
        await api.get<Folder[]>(`/spaces/${spaceId}/folders`, { params }),
      );
    },
  });

export const folderQueryOptions = (id: number) =>
  queryOptions({
    queryKey: folderKeys.detail(id),
    queryFn: async () =>
      folderSchema.parse(await api.get<Folder>(`/folders/${id}`)),
  });

export const breadcrumbQueryOptions = (id: number) =>
  queryOptions({
    queryKey: folderKeys.breadcrumb(id),
    queryFn: async () =>
      breadcrumbListSchema.parse(
        await api.get<BreadcrumbItem[]>(`/folders/${id}/breadcrumb`),
      ),
  });

export const folderChildrenQueryOptions = (id: number) =>
  queryOptions({
    queryKey: folderKeys.children(id),
    queryFn: async () =>
      folderListSchema.parse(
        await api.get<Folder[]>(`/folders/${id}/children`),
      ),
  });

export function useFolders(spaceId: number, parentId?: number | null) {
  return useQuery(foldersQueryOptions(spaceId, parentId));
}

export function useFolder(id: number) {
  return useQuery(folderQueryOptions(id));
}

export function useBreadcrumb(id: number) {
  return useQuery(breadcrumbQueryOptions(id));
}

export function useFolderChildren(id: number) {
  return useQuery(folderChildrenQueryOptions(id));
}

export function useCreateFolder(spaceId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateFolderInput) =>
      api.post<Folder>(`/spaces/${spaceId}/folders`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: folderKeys.lists() });
    },
  });
}

export function useRenameFolder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateFolderInput }) =>
      api.patch<Folder>(`/folders/${id}`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: folderKeys.all });
    },
  });
}

export function useMoveFolder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: MoveFolderInput }) =>
      api.post<Folder>(`/folders/${id}/move`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: folderKeys.all });
    },
  });
}

export function useDeleteFolder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.delete<void>(`/folders/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: folderKeys.lists() });
    },
  });
}

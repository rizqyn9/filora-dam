import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import {
  gallListSchema,
  gallerySchema,
  type CreateGalleryInput,
  type Gallery,
  type UpdateGalleryInput,
} from "@/features/galleries/schemas";
import { api } from "@/lib/api-client";

export const galleryKeys = {
  all: ["galleries"] as const,
  lists: () => [...galleryKeys.all, "list"] as const,
  list: () => [...galleryKeys.lists()] as const,
  details: () => [...galleryKeys.all, "detail"] as const,
  detail: (id: number) => [...galleryKeys.details(), id] as const,
};

export const galleriesQueryOptions = () =>
  queryOptions({
    queryKey: galleryKeys.list(),
    queryFn: async () =>
      gallListSchema.parse(await api.get<Gallery[]>("/galleries")),
  });

export const galleryQueryOptions = (id: number) =>
  queryOptions({
    queryKey: galleryKeys.detail(id),
    queryFn: async () =>
      gallerySchema.parse(await api.get<Gallery>(`/galleries/${id}`)),
  });

export function useGalleries() {
  return useQuery(galleriesQueryOptions());
}

export function useGallery(id: number) {
  return useQuery(galleryQueryOptions(id));
}

export function useCreateGallery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateGalleryInput) =>
      api.post<Gallery>("/galleries", input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: galleryKeys.lists() });
    },
  });
}

export function useUpdateGallery(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateGalleryInput) =>
      api.patch<Gallery>(`/galleries/${id}`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: galleryKeys.all });
    },
  });
}

export function useDeleteGallery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.delete<void>(`/galleries/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: galleryKeys.lists() });
    },
  });
}

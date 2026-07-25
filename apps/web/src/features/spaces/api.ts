import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import {
  spaceListSchema,
  spaceSchema,
  type CreateSpaceInput,
  type Space,
  type UpdateSpaceInput,
} from "@/features/spaces/schemas";
import { api } from "@/lib/api-client";

export const spaceKeys = {
  all: ["spaces"] as const,
  lists: () => [...spaceKeys.all, "list"] as const,
  list: () => [...spaceKeys.lists()] as const,
  details: () => [...spaceKeys.all, "detail"] as const,
  detail: (id: number) => [...spaceKeys.details(), id] as const,
};

export const spacesQueryOptions = () =>
  queryOptions({
    queryKey: spaceKeys.list(),
    queryFn: async () =>
      spaceListSchema.parse(await api.get<Space[]>("/spaces")),
  });

export const spaceQueryOptions = (id: number) =>
  queryOptions({
    queryKey: spaceKeys.detail(id),
    queryFn: async () =>
      spaceSchema.parse(await api.get<Space>(`/spaces/${id}`)),
  });

export function useSpaces() {
  return useQuery(spacesQueryOptions());
}

export function useSpace(id: number) {
  return useQuery(spaceQueryOptions(id));
}

export function useCreateSpace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateSpaceInput) =>
      api.post<Space>("/spaces", input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: spaceKeys.lists() });
    },
  });
}

export function useUpdateSpace(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateSpaceInput) =>
      api.patch<Space>(`/spaces/${id}`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: spaceKeys.all });
    },
  });
}

export function useDeleteSpace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.delete<void>(`/spaces/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: spaceKeys.lists() });
    },
  });
}

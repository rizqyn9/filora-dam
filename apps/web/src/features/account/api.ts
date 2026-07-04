import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import {
  userSchema,
  type UpdateProfileInput,
  type User,
} from "@/features/account/schemas";
import { api } from "@/lib/api-client";

export const accountKeys = {
  all: ["account"] as const,
  me: () => [...accountKeys.all, "me"] as const,
};

export const meQueryOptions = () =>
  queryOptions({
    queryKey: accountKeys.me(),
    queryFn: async () => userSchema.parse(await api.get<User>("/me")),
  });

export function useMe() {
  return useQuery(meQueryOptions());
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateProfileInput) => api.patch<User>("/me", input),
    onSuccess: (user) => {
      queryClient.setQueryData(accountKeys.me(), userSchema.parse(user));
    },
  });
}

import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";

import { SpaceSchema, type Space } from "./schemas";

export function useSpaces() {
  return useQuery<Space[]>({
    queryKey: ["spaces"],
    queryFn: async () => {
      const raw = await api<unknown[]>("/spaces");
      return raw.map((s) => SpaceSchema.parse(s));
    },
  });
}

import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useCreateSpace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) =>
      api<Space>("/spaces", {
        method: "POST",
        body: JSON.stringify({ name }),
      }).then((raw) => SpaceSchema.parse(raw)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["spaces"] });
    },
  });
}

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

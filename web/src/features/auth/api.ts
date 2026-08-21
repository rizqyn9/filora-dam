import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { api } from "@/lib/api";

import { useAuthStore } from "./auth-store";
import {
  AuthResponseSchema,
  type AuthResponse,
  type User,
  UserSchema,
} from "./schemas";

export function useLogin() {
  const setToken = useAuthStore((s) => s.setToken);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (input: { email: string; password: string }) => {
      const raw = await api<AuthResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ ...input, client: "web" }),
      });
      return AuthResponseSchema.parse(raw);
    },
    onSuccess: (data) => {
      setToken(data.token);
      navigate({ to: "/" });
    },
  });
}

export function useLogout() {
  const clearToken = useAuthStore((s) => s.clearToken);
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: () => api("/auth/logout", { method: "POST" }),
    onSettled: () => {
      clearToken();
      queryClient.clear();
      navigate({ to: "/login" });
    },
  });
}

export function useMe() {
  const token = useAuthStore((s) => s.token);

  return useQuery<User>({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      const raw = await api<User>("/auth/me");
      return UserSchema.parse(raw);
    },
    enabled: !!token,
    staleTime: 1000 * 60 * 5, // 5 min
    retry: false,
  });
}

import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/features/auth/auth-store";
import { AuthResponseSchema } from "@/features/auth/schemas";
import { api } from "@/lib/api";
import { useMutation } from "@tanstack/react-query";

export const Route = createFileRoute("/register")({
  beforeLoad: () => {
    if (useAuthStore.getState().token) {
      throw redirect({ to: "/" });
    }
  },
  component: RegisterPage,
});

function RegisterPage() {
  const searchParams = new URLSearchParams(window.location.search);
  const inviteToken = searchParams.get("token") ?? "";

  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const setToken = useAuthStore((s) => s.setToken);
  const navigate = useNavigate();

  const register = useMutation({
    mutationFn: async () => {
      const raw = await api("/auth/register", {
        method: "POST",
        body: JSON.stringify({ invite_token: inviteToken, name, password }),
      });
      return AuthResponseSchema.parse(raw);
    },
    onSuccess: (data) => {
      setToken(data.token);
      navigate({ to: "/" });
    },
  });

  if (!inviteToken) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <p className="text-muted-foreground">
          Invalid invite link. Please use the link you received.
        </p>
      </div>
    );
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    register.mutate();
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="space-y-2 text-center">
          <h1 className="text-2xl font-semibold">Join Filora</h1>
          <p className="text-muted-foreground text-sm">Set up your account</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              required
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              minLength={8}
            />
          </div>

          {register.error && (
            <p className="text-destructive text-sm">{register.error.message}</p>
          )}

          <Button type="submit" className="w-full" disabled={register.isPending}>
            {register.isPending ? "Creating account..." : "Create account"}
          </Button>
        </form>
      </div>
    </div>
  );
}

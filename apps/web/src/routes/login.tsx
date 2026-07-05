import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Boxes, Loader2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";

export const Route = createFileRoute("/login")({
  component: LoginPage,
});

function LoginPage() {
  const navigate = useNavigate();
  const setToken = useAuthStore((s) => s.setToken);
  const [loading, setLoading] = useState(false);

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    // Placeholder auth — swap for Clerk sign-in. For now we mint a dev token so
    // the app shell becomes reachable.
    setTimeout(() => {
      setToken("dev-token");
      toast.success("Welcome back");
      void navigate({ to: "/" });
    }, 600);
  };

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Brand panel */}
      <div className="relative hidden flex-col justify-between bg-primary p-10 text-primary-foreground lg:flex">
        <div className="flex items-center gap-2 font-semibold">
          <div className="flex size-8 items-center justify-center rounded-md bg-primary-foreground/15">
            <Boxes className="size-5" />
          </div>
          Filora
        </div>
        <div className="space-y-3">
          <p className="text-2xl leading-snug font-semibold">
            Every asset, everywhere — organized in one place.
          </p>
          <p className="text-sm text-primary-foreground/70">
            Multi-cloud digital asset management. Storage complexity, abstracted
            away.
          </p>
        </div>
        <p className="text-xs text-primary-foreground/50">
          © {new Date().getFullYear()} Filora
        </p>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center p-6">
        <Card className="w-full max-w-sm border-0 shadow-none sm:border sm:shadow-sm">
          <CardHeader className="space-y-1">
            <CardTitle className="text-xl">Sign in</CardTitle>
            <CardDescription>
              Enter your credentials to access your workspace.
            </CardDescription>
          </CardHeader>
          <form onSubmit={onSubmit}>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@filora.app"
                  autoComplete="email"
                  required
                />
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password">Password</Label>
                  <button
                    type="button"
                    className="text-xs text-muted-foreground hover:text-foreground"
                  >
                    Forgot password?
                  </button>
                </div>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  autoComplete="current-password"
                  required
                />
              </div>
            </CardContent>
            <CardFooter className="mt-6 flex-col gap-3">
              <Button type="submit" className="w-full" disabled={loading}>
                {loading && <Loader2 className="size-4 animate-spin" />}
                Sign in
              </Button>
              <p className="text-center text-xs text-muted-foreground">
                Don&apos;t have an account?{" "}
                <button
                  type="button"
                  className="font-medium text-foreground hover:underline"
                >
                  Request access
                </button>
              </p>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

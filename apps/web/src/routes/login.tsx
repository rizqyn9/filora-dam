import { SignIn } from "@clerk/clerk-react";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { Boxes } from "lucide-react";

import { useAuthStore } from "@/stores/auth-store";

export const Route = createFileRoute("/login")({
  beforeLoad: () => {
    if (useAuthStore.getState().isAuthenticated) {
      throw redirect({ to: "/" });
    }
  },
  component: LoginPage,
});

function LoginPage() {
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

      {/* Clerk sign-in */}
      <div className="flex items-center justify-center p-6">
        <SignIn
          routing="hash"
          signUpUrl="/login"
          forceRedirectUrl="/"
          appearance={{ elements: { rootBox: "w-full max-w-sm" } }}
        />
      </div>
    </div>
  );
}

import type { QueryClient } from "@tanstack/react-query";
import {
  createRootRouteWithContext,
  Link,
  Outlet,
} from "@tanstack/react-router";
import { lazy, Suspense } from "react";

import { Toaster } from "@/components/ui/sonner";

/** Router context — available in every route's loader / beforeLoad. */
export interface RouterContext {
  queryClient: QueryClient;
}

// Devtools are dev-only and code-split out of the production bundle.
const TanStackRouterDevtools = import.meta.env.PROD
  ? () => null
  : lazy(() =>
      import("@tanstack/react-router-devtools").then((m) => ({
        default: m.TanStackRouterDevtools,
      })),
    );

function RootLayout() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <nav className="mx-auto flex h-14 max-w-6xl items-center gap-6 px-4">
          <Link to="/" className="font-semibold">
            Filora
          </Link>
          <Link
            to="/galleries"
            className="text-sm text-muted-foreground [&.active]:text-foreground"
          >
            Galleries
          </Link>
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-8">
        <Outlet />
      </main>
      <Toaster />
      <Suspense>
        <TanStackRouterDevtools />
      </Suspense>
    </div>
  );
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
});

import type { QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import { lazy, Suspense } from "react";

import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useApplyTheme } from "@/hooks/use-theme";

export interface RouterContext {
  queryClient: QueryClient;
}

const TanStackRouterDevtools = import.meta.env.PROD
  ? () => null
  : lazy(() =>
      import("@tanstack/react-router-devtools").then((m) => ({
        default: m.TanStackRouterDevtools,
      })),
    );

function RootLayout() {
  useApplyTheme();

  return (
    <TooltipProvider>
      <Outlet />
      <Toaster />
      <Suspense>
        <TanStackRouterDevtools />
      </Suspense>
    </TooltipProvider>
  );
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
});

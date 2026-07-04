import { createRouter } from "@tanstack/react-router";

import { queryClient } from "@/lib/query-client";
import { routeTree } from "@/routeTree.gen";

/**
 * SPA router. Route context carries the shared queryClient so route loaders
 * can prefetch server state (see routes/galleries/index.tsx).
 */
export const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: "intent",
  // React Query owns caching; the router shouldn't cache loader data itself.
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

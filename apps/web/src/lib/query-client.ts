import { QueryClient } from "@tanstack/react-query";

import { ApiError } from "@/lib/api-client";

/**
 * Shared TanStack Query client.
 *
 * Server state lives here (assets, galleries, albums, ...). Client/UI state
 * lives in Zustand stores (see src/stores). Keep the two separate.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failureCount, error) => {
        // Don't retry auth/permission/not-found errors.
        if (
          error instanceof ApiError &&
          [401, 403, 404].includes(error.status)
        ) {
          return false;
        }
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: false,
    },
  },
});

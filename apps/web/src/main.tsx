import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { StrictMode, lazy, Suspense } from "react";
import { createRoot } from "react-dom/client";

import { queryClient } from "@/lib/query-client";
import { router } from "@/router";
import { initAuthBridge } from "@/stores/auth-store";

import "@/styles.css";

// Bind the API client's token provider + 401 handler to the auth store.
initAuthBridge();

const ReactQueryDevtools = import.meta.env.PROD
  ? () => null
  : lazy(() =>
      import("@tanstack/react-query-devtools").then((m) => ({
        default: m.ReactQueryDevtools,
      })),
    );

const rootElement = document.getElementById("root");
if (!rootElement) throw new Error("Root element #root not found");

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Suspense>
        <ReactQueryDevtools initialIsOpen={false} />
      </Suspense>
    </QueryClientProvider>
  </StrictMode>,
);

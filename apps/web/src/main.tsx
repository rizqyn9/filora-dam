import { ClerkProvider } from "@clerk/clerk-react";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { StrictMode, lazy, Suspense } from "react";
import { createRoot } from "react-dom/client";

import { ClerkAuthBridge } from "@/components/clerk-auth-bridge";
import { env } from "@/lib/env";
import { queryClient } from "@/lib/query-client";
import { router } from "@/router";

import "@/styles.css";

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
    <ClerkProvider
      publishableKey={env.VITE_CLERK_PUBLISHABLE_KEY}
      afterSignOutUrl="/login"
    >
      <QueryClientProvider client={queryClient}>
        <ClerkAuthBridge>
          <RouterProvider router={router} />
        </ClerkAuthBridge>
        <Suspense>
          <ReactQueryDevtools initialIsOpen={false} />
        </Suspense>
      </QueryClientProvider>
    </ClerkProvider>
  </StrictMode>,
);

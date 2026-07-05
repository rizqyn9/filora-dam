import { useAuth, useClerk } from "@clerk/clerk-react";
import type { ReactNode } from "react";

import {
  setAuthTokenProvider,
  setUnauthorizedHandler,
} from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Bridges Clerk into the app's auth-agnostic plumbing:
 *
 *  - Feeds the api client a fresh Clerk session token on every request
 *    (Clerk auto-refreshes it), and signs the user out on a 401.
 *  - Mirrors Clerk's `isSignedIn` into the zustand store so the `_app` route
 *    guard (which runs outside React) can read it synchronously.
 *
 * Rendering is gated on Clerk being loaded so the router's `beforeLoad` guard
 * never sees a stale "signed out" state on a hard refresh.
 */
export function ClerkAuthBridge({ children }: { children: ReactNode }) {
  const { isLoaded, isSignedIn, getToken } = useAuth();
  const { signOut } = useClerk();

  // Wire the api client to Clerk. Assigning during render is idempotent (just
  // swaps function refs) and guarantees the provider is set before the router
  // — and any query it kicks off — mounts below.
  setAuthTokenProvider(() => getToken());
  setUnauthorizedHandler(() => void signOut());

  if (!isLoaded) {
    return (
      <div className="grid min-h-screen place-items-center text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }

  // Keep the guard's view of auth in sync before children mount. Guarded so we
  // don't schedule redundant updates on every render.
  if (useAuthStore.getState().isAuthenticated !== isSignedIn) {
    useAuthStore.setState({ isAuthenticated: isSignedIn });
  }

  return children;
}

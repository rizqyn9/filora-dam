import { create } from "zustand";

import { setAuthTokenProvider, setUnauthorizedHandler } from "@/lib/api-client";

/**
 * Auth session state — auth-agnostic.
 *
 * Today it holds a bearer token directly (dev / CLI-token flows). When Clerk is
 * wired in, replace `setToken` usage with a Clerk token provider:
 *
 *   setAuthTokenProvider(() => clerk.session?.getToken() ?? null)
 *
 * ...and this store can just mirror `isAuthenticated`. Nothing else in the app
 * needs to change because everything reads auth through this store / the api
 * client's injected provider.
 */
interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  setToken: (token: string | null) => void;
  clear: () => void;
}

const STORAGE_KEY = "filora-auth-token";

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem(STORAGE_KEY),
  isAuthenticated: Boolean(localStorage.getItem(STORAGE_KEY)),

  setToken: (token) => {
    if (token) localStorage.setItem(STORAGE_KEY, token);
    else localStorage.removeItem(STORAGE_KEY);
    set({ token, isAuthenticated: Boolean(token) });
  },

  clear: () => {
    localStorage.removeItem(STORAGE_KEY);
    set({ token: null, isAuthenticated: false });
  },
}));

/**
 * Bind the api client's token provider + 401 handler to this store.
 * Call once at bootstrap (see main.tsx).
 */
export function initAuthBridge() {
  setAuthTokenProvider(() => useAuthStore.getState().token);
  setUnauthorizedHandler(() => useAuthStore.getState().clear());
}

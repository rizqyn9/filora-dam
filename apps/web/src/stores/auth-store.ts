import { create } from "zustand";

/**
 * Auth session state — a Clerk-agnostic mirror of "is the user signed in".
 *
 * Clerk owns the real session and token lifecycle (see components/clerk-auth-
 * bridge.tsx, which pushes Clerk state in here and wires the api client's token
 * provider). This store exists only so non-React call sites — notably the
 * `_app` route guard's `beforeLoad`, which runs outside React — can read auth
 * state synchronously via `useAuthStore.getState()`.
 */
interface AuthState {
  isAuthenticated: boolean;
  setAuthenticated: (value: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  setAuthenticated: (value) => set({ isAuthenticated: value }),
}));

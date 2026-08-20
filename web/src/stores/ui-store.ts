import { create } from "zustand";
import { persist } from "zustand/middleware";

type Theme = "light" | "dark" | "system";

interface UiState {
  theme: Theme;
  sidebarOpen: boolean;
  /** Multi-select of asset ids in the current view (client-only UI state). */
  selectedAssetIds: Set<string>;

  setTheme: (theme: Theme) => void;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  toggleAssetSelection: (id: string) => void;
  clearAssetSelection: () => void;
}

/**
 * Global UI/client state. NOT server data — that belongs in TanStack Query.
 */
export const useUiStore = create<UiState>()(
  persist(
    (set) => ({
      theme: "system",
      sidebarOpen: true,
      selectedAssetIds: new Set<string>(),

      setTheme: (theme) => set({ theme }),
      toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
      setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
      toggleAssetSelection: (id) =>
        set((s) => {
          const next = new Set(s.selectedAssetIds);
          if (next.has(id)) {
            next.delete(id);
          } else {
            next.add(id);
          }
          return { selectedAssetIds: next };
        }),
      clearAssetSelection: () => set({ selectedAssetIds: new Set<string>() }),
    }),
    {
      name: "filora-ui",
      // Only persist durable preferences, not transient selection.
      partialize: (state) => ({
        theme: state.theme,
        sidebarOpen: state.sidebarOpen,
      }),
    },
  ),
);

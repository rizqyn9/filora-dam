import { create } from "zustand";
import { persist } from "zustand/middleware";

type Theme = "light" | "dark" | "system";
type ViewMode = "grid" | "list";

interface UiState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  viewMode: ViewMode;
  setViewMode: (mode: ViewMode) => void;
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
}

export const useUiStore = create<UiState>()(
  persist(
    (set) => ({
      theme: "system",
      setTheme: (theme) => set({ theme }),
      viewMode: "grid",
      setViewMode: (viewMode) => set({ viewMode }),
      sidebarOpen: false,
      setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
    }),
    {
      name: "filora-ui",
      partialize: (state) => ({ theme: state.theme, viewMode: state.viewMode }),
    },
  ),
);

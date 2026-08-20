import { useEffect } from "react";

import { useUiStore } from "@/stores/ui-store";

/**
 * Applies the selected theme to <html> and keeps it in sync with the OS when
 * set to "system". Call once near the app root.
 */
export function useApplyTheme() {
  const theme = useUiStore((s) => s.theme);

  useEffect(() => {
    const root = document.documentElement;
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    const apply = () => {
      const isDark = theme === "dark" || (theme === "system" && media.matches);
      root.classList.toggle("dark", isDark);
    };

    apply();
    if (theme === "system") {
      media.addEventListener("change", apply);
      return () => media.removeEventListener("change", apply);
    }
  }, [theme]);
}

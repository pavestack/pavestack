export type Theme = "light" | "dark";

const STORAGE_KEY = "pavestack-portal:theme";

function systemPrefersLight(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) return false;
  return window.matchMedia("(prefers-color-scheme: light)").matches;
}

/** Resolve the theme to use on first paint: explicit choice, else OS preference, else dark. */
export function getInitialTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark") return stored;
  return systemPrefersLight() ? "light" : "dark";
}

/** Apply a theme to <html data-theme="..."> and persist the explicit choice. */
export function applyTheme(theme: Theme): void {
  if (typeof document === "undefined") return;
  document.documentElement.setAttribute("data-theme", theme);
  try {
    window.localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    // localStorage unavailable (private mode / disabled) — theme still applies for this session
  }
}

export function toggleTheme(current: Theme): Theme {
  return current === "dark" ? "light" : "dark";
}

export const THEME_STORAGE_KEY = STORAGE_KEY;

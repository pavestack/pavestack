import { describe, test, expect, beforeEach, afterEach, vi } from "vitest";
import { applyTheme, getInitialTheme, toggleTheme, THEME_STORAGE_KEY } from "./theme";

describe("theme persistence", () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
  });

  afterEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
  });

  test("toggleTheme flips between light and dark", () => {
    expect(toggleTheme("dark")).toBe("light");
    expect(toggleTheme("light")).toBe("dark");
  });

  test("getInitialTheme reads an explicit stored choice over OS preference", () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, "light");
    vi.spyOn(window, "matchMedia").mockReturnValue({ matches: false } as MediaQueryList);
    expect(getInitialTheme()).toBe("light");
  });

  test("getInitialTheme falls back to OS preference when nothing is stored", () => {
    vi.spyOn(window, "matchMedia").mockImplementation(
      (query: string) => ({ matches: query.includes("light") } as MediaQueryList)
    );
    expect(getInitialTheme()).toBe("light");
  });

  test("getInitialTheme defaults to dark when no preference and no OS match", () => {
    vi.spyOn(window, "matchMedia").mockReturnValue({ matches: false } as MediaQueryList);
    expect(getInitialTheme()).toBe("dark");
  });

  test("applyTheme sets the data-theme attribute and persists to localStorage", () => {
    applyTheme("light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe("light");

    applyTheme("dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe("dark");
  });
});

import React, { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCatalog } from "../CatalogContext";
import { useTheme } from "../ThemeContext";
import { IconMenu, IconMoon, IconSearch, IconSun } from "../icons";

export function Header({ onToggleSidebar }: { onToggleSidebar: () => void }) {
  const { theme, toggle } = useTheme();
  const { catalog } = useCatalog();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [focused, setFocused] = useState(false);

  const suggestions = useMemo(() => {
    if (!query.trim() || !catalog) return [];
    const lower = query.toLowerCase();
    return catalog.services.filter((s) => s.name.toLowerCase().includes(lower)).slice(0, 6);
  }, [query, catalog]);

  function goToService(name: string) {
    setQuery("");
    setFocused(false);
    navigate(`/services/${name}`);
  }

  return (
    <header className="app-header sticky top-0 z-10">
      <div className="flex items-center gap-3 px-4 h-14">
        <button
          type="button"
          aria-label="Toggle navigation"
          onClick={onToggleSidebar}
          className="md:hidden btn btn-secondary !px-2 !py-2"
        >
          <IconMenu />
        </button>

        <div className="relative flex-1 max-w-md">
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-pave-text-muted pointer-events-none">
            <IconSearch />
          </div>
          <input
            id="global-search"
            type="search"
            role="combobox"
            aria-expanded={focused && suggestions.length > 0}
            aria-controls="global-search-results"
            aria-label="Search services"
            placeholder="Search services…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setFocused(true)}
            onBlur={() => setTimeout(() => setFocused(false), 120)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && suggestions[0]) goToService(suggestions[0].name);
            }}
            className="search-input pl-10"
          />
          {focused && suggestions.length > 0 && (
            <ul
              id="global-search-results"
              role="listbox"
              className="absolute z-20 mt-1 w-full rounded-lg border border-pave-border bg-pave-elevated shadow-card overflow-hidden"
            >
              {suggestions.map((s) => (
                <li key={s.id} role="option" aria-selected="false">
                  <button
                    type="button"
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => goToService(s.name)}
                    className="w-full text-left px-3 py-2 text-sm text-pave-text hover:bg-pave-surface-hover"
                  >
                    <span className="font-medium">{s.name}</span>
                    <span className="text-pave-text-muted ml-2 text-xs">{s.team}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="ml-auto flex items-center gap-2">
          <button
            type="button"
            aria-label={theme === "dark" ? "Switch to light theme" : "Switch to dark theme"}
            onClick={toggle}
            className="btn btn-secondary !px-2 !py-2"
          >
            {theme === "dark" ? <IconSun /> : <IconMoon />}
          </button>
        </div>
      </div>
    </header>
  );
}

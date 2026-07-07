import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { Catalog } from "../lib/catalog";

type CatalogContextValue = {
  catalog: Catalog | null;
  loading: boolean;
  error: boolean;
  reload: () => void;
};

const CatalogContext = createContext<CatalogContextValue | null>(null);

export function CatalogProvider({ children }: { children: React.ReactNode }) {
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [nonce, setNonce] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);
    fetch("/catalog.json")
      .then((response) => {
        if (!response.ok) throw new Error("failed to load catalog");
        return response.json() as Promise<Catalog>;
      })
      .then((data) => {
        if (cancelled) return;
        setCatalog(data);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setError(true);
        setCatalog({ generatedAt: new Date(0).toISOString(), services: [] });
        setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [nonce]);

  const reload = useCallback(() => setNonce((n) => n + 1), []);

  return (
    <CatalogContext.Provider value={{ catalog, loading, error, reload }}>
      {children}
    </CatalogContext.Provider>
  );
}

export function useCatalog(): CatalogContextValue {
  const ctx = useContext(CatalogContext);
  if (!ctx) throw new Error("useCatalog must be used within a CatalogProvider");
  return ctx;
}

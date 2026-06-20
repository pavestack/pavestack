import React, { useEffect, useMemo, useState } from "react";
import type { Catalog, SortKey } from "../lib/catalog";
import {
  filterServices,
  sortServices,
  computeStats,
} from "../lib/catalog";
import {
  StatCard,
  ServiceCard,
  SortSelect,
} from "./components";
import {
  IconServices,
  IconScore,
  IconPassing,
  IconCheck,
  IconSearch,
} from "./icons";

export function App() {
  const [data, setData] = useState<Catalog | null>(null);
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("name");
  const [error, setError] = useState(false);

  useEffect(() => {
    fetch("/catalog.json")
      .then((response) => {
        if (!response.ok) throw new Error("failed to load catalog");
        return response.json() as Promise<Catalog>;
      })
      .then(setData)
      .catch(() => {
        setError(true);
        setData({ generatedAt: new Date(0).toISOString(), services: [] });
      });
  }, []);

  const filtered = useMemo(() => {
    if (!data) return [];
    return sortServices(filterServices(data.services, search), sortKey);
  }, [data, search, sortKey]);

  const stats = useMemo(() => {
    if (!data) return null;
    return computeStats(data.services);
  }, [data]);

  /* ── Loading state ── */
  if (!data) {
    return (
      <main className="min-h-screen flex items-center justify-center">
        <div className="flex items-center gap-3 text-pave-500">
          <div className="w-5 h-5 border-2 border-pave-700 border-t-pave-accent rounded-full animate-spin" />
          Loading catalog…
        </div>
      </main>
    );
  }

  return (
    <div className="min-h-screen">
      {/* ── Hero Header ── */}
      <header className="border-b border-pave-800 bg-pave-950/80 backdrop-blur-md sticky top-0 z-10">
        <div className="mx-auto max-w-7xl px-6 py-5">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="animate-fade-in">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-pave-accent font-semibold text-sm tracking-wide uppercase">
                  Pavestack
                </span>
                <span className="badge badge-neutral text-[10px]">IDP</span>
              </div>
              <h1 className="text-2xl sm:text-3xl font-bold text-pave-300">
                Software Catalog
              </h1>
            </div>
            <div className="text-right animate-fade-in" style={{ animationDelay: "0.1s" }}>
              <p className="text-xs text-pave-600">
                Last generated
              </p>
              <p className="font-mono text-xs text-pave-500">
                {new Date(data.generatedAt).toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-6 py-8">
        {/* ── Stats Row ── */}
        {stats && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard
              icon={<IconServices />}
              label="Services"
              value={stats.total}
              subtext="Registered in catalog"
              delay={0.05}
            />
            <StatCard
              icon={<IconScore />}
              label="Avg Score"
              value={stats.avgScore}
              subtext="Platform compliance"
              delay={0.1}
            />
            <StatCard
              icon={<IconPassing />}
              label="Passing"
              value={`${stats.passing}/${stats.total}`}
              subtext="Score ≥ 70"
              delay={0.15}
            />
            <StatCard
              icon={<IconCheck />}
              label="Criteria"
              value={`${stats.passingCriteria}/${stats.totalCriteria}`}
              subtext="Individual checks"
              delay={0.2}
            />
          </div>
        )}

        {/* ── Search & Sort Bar ── */}
        <div className="flex flex-col sm:flex-row gap-3 mb-6 animate-slide-up" style={{ animationDelay: "0.15s" }}>
          <div className="relative flex-1">
            <div className="absolute left-3 top-1/2 -translate-y-1/2">
              <IconSearch />
            </div>
            <input
              id="search-services"
              type="text"
              placeholder="Search services by name, owner, or description…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="search-input pl-10"
            />
          </div>
          <SortSelect value={sortKey} onChange={setSortKey} />
        </div>

        {/* ── Error Banner ── */}
        {error && (
          <div className="rounded-lg border border-pave-warning/30 bg-pave-warning/5 px-4 py-3 mb-6 text-sm text-pave-warning animate-fade-in">
            ⚠ Failed to load catalog data. Showing empty state.
          </div>
        )}

        {/* ── Service Grid ── */}
        <section className="grid gap-5 md:grid-cols-2">
          {filtered.map((service, i) => (
            <ServiceCard key={service.id} service={service} index={i} />
          ))}
        </section>

        {/* ── Empty States ── */}
        {filtered.length === 0 && data.services.length > 0 && (
          <div className="text-center py-16 animate-fade-in">
            <p className="text-pave-500 text-lg mb-2">No services match your search</p>
            <p className="text-pave-600 text-sm">
              Try a different search term or{" "}
              <button onClick={() => setSearch("")} className="text-pave-accent hover:underline">
                clear the filter
              </button>
            </p>
          </div>
        )}

        {data.services.length === 0 && (
          <div className="text-center py-16 animate-fade-in">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-pave-800 mb-4">
              <IconServices />
            </div>
            <p className="text-pave-400 text-lg mb-2">No services registered</p>
            <p className="text-pave-600 text-sm">
              Run <code className="font-mono bg-pave-800 px-2 py-0.5 rounded text-pave-accent">pave create-service</code> to scaffold your first service.
            </p>
          </div>
        )}

        {/* ── Footer ── */}
        <footer className="mt-16 pt-6 border-t border-pave-800 text-center">
          <p className="text-xs text-pave-600">
            Pavestack Portal — Read-only visibility layer · Does not mutate state
          </p>
        </footer>
      </main>
    </div>
  );
}

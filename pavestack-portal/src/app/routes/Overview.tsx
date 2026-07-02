import React, { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useCatalog } from "../CatalogContext";
import {
  computeStats,
  filterByTeam,
  filterByTier,
  filterServices,
  generateDemoServices,
  sortServices,
  type SortKey,
} from "../../lib/catalog";
import { DemoDataNote, EmptyState, InlineError, ServiceCard, SkeletonGrid, SortSelect, StatCard } from "../components";
import { ServiceDataTable, sortServicesForTable, type SortDir, type TableSortKey } from "../DataTable";
import { IconCheck, IconGrid, IconPassing, IconScore, IconServices } from "../icons";

type ViewMode = "grid" | "table";

export function Overview() {
  const { catalog, loading, error, reload } = useCatalog();
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("name");
  const [team, setTeam] = useState("");
  const [tier, setTier] = useState("");
  const [view, setView] = useState<ViewMode>("grid");
  const [showDemo, setShowDemo] = useState(false);
  const [tableSortKey, setTableSortKey] = useState<TableSortKey>("name");
  const [tableSortDir, setTableSortDir] = useState<SortDir>("asc");

  const demoServices = useMemo(() => (showDemo ? generateDemoServices(1200) : []), [showDemo]);

  const allServices = useMemo(() => {
    const real = catalog?.services ?? [];
    return showDemo ? [...real, ...demoServices] : real;
  }, [catalog, showDemo, demoServices]);

  const teams = useMemo(() => Array.from(new Set(allServices.map((s) => s.team))).sort(), [allServices]);
  const tiers = ["tier-1", "tier-2", "tier-3"];

  const filtered = useMemo(() => {
    let result = filterServices(allServices, search);
    result = filterByTeam(result, team);
    result = filterByTier(result, tier);
    return result;
  }, [allServices, search, team, tier]);

  const gridSorted = useMemo(() => sortServices(filtered, sortKey), [filtered, sortKey]);
  const tableSorted = useMemo(() => sortServicesForTable(filtered, tableSortKey, tableSortDir), [filtered, tableSortKey, tableSortDir]);

  const stats = useMemo(() => (catalog ? computeStats(catalog.services) : null), [catalog]);

  function handleTableSort(key: TableSortKey) {
    if (key === tableSortKey) {
      setTableSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setTableSortKey(key);
      setTableSortDir("asc");
    }
  }

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
        <div>
          <h1 className="text-xl font-semibold text-pave-text">Service catalog</h1>
          <p className="text-sm text-pave-text-muted mt-0.5">
            Generated from committed <code className="font-mono">catalog-info.yaml</code> and{" "}
            <code className="font-mono">scorecard.yaml</code> files
            {catalog && <> · last generated {new Date(catalog.generatedAt).toLocaleString()}</>}
          </p>
        </div>
      </div>

      {loading && <SkeletonGrid rows={4} />}

      {!loading && error && (
        <div className="mb-6">
          <InlineError message="Failed to load catalog.json. Showing an empty catalog below." onRetry={reload} />
        </div>
      )}

      {!loading && catalog && (
        <>
          {stats && (
            <div className="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-6">
              <StatCard icon={<IconServices />} label="Services" value={stats.total} subtext="Registered in catalog" />
              <StatCard icon={<IconScore />} label="Avg score" value={stats.avgScore} subtext="Platform compliance" />
              <StatCard icon={<IconPassing />} label="Passing" value={`${stats.passing}/${stats.total}`} subtext="Score ≥ 70" />
              <StatCard icon={<IconCheck />} label="Criteria" value={`${stats.passingCriteria}/${stats.totalCriteria}`} subtext="Individual checks" />
              <StatCard
                icon={<IconGrid />}
                label="Pave adoption"
                value={`${stats.adoptionPct}%`}
                subtext={`${stats.createdViaPave}/${stats.total} created via pave-cli`}
              />
            </div>
          )}

          <div className="flex flex-col lg:flex-row gap-3 mb-4">
            <div className="relative flex-1">
              <input
                id="search-services"
                type="text"
                placeholder="Search services by name, owner, or description…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="search-input"
              />
            </div>
            <select
              aria-label="Filter by team"
              value={team}
              onChange={(e) => setTeam(e.target.value)}
              className="rounded-lg border border-pave-border bg-pave-surface px-3 py-2 text-sm text-pave-text-secondary outline-none hover:border-pave-border-strong"
            >
              <option value="">All teams</option>
              {teams.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            <select
              aria-label="Filter by tier"
              value={tier}
              onChange={(e) => setTier(e.target.value)}
              className="rounded-lg border border-pave-border bg-pave-surface px-3 py-2 text-sm text-pave-text-secondary outline-none hover:border-pave-border-strong"
            >
              <option value="">All tiers</option>
              {tiers.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            {view === "grid" && <SortSelect value={sortKey} onChange={setSortKey} />}
            <div className="flex rounded-lg border border-pave-border overflow-hidden shrink-0" role="group" aria-label="View mode">
              <button
                type="button"
                onClick={() => setView("grid")}
                aria-pressed={view === "grid"}
                className={`px-3 py-2 text-sm ${view === "grid" ? "bg-pave-accent/10 text-pave-accent" : "text-pave-text-secondary hover:bg-pave-surface-hover"}`}
              >
                Cards
              </button>
              <button
                type="button"
                onClick={() => setView("table")}
                aria-pressed={view === "table"}
                className={`px-3 py-2 text-sm border-l border-pave-border ${view === "table" ? "bg-pave-accent/10 text-pave-accent" : "text-pave-text-secondary hover:bg-pave-surface-hover"}`}
              >
                Table
              </button>
            </div>
          </div>

          <div className="flex items-center gap-2 mb-4">
            <label className="flex items-center gap-2 text-xs text-pave-text-muted cursor-pointer select-none">
              <input type="checkbox" checked={showDemo} onChange={(e) => setShowDemo(e.target.checked)} className="accent-[var(--accent)]" />
              Show 1,200 synthetic demo rows (table scale/virtualization demo)
            </label>
          </div>

          {showDemo && (
            <div className="mb-4">
              <DemoDataNote count={demoServices.length} />
            </div>
          )}

          {filtered.length === 0 && allServices.length > 0 && (
            <EmptyState
              title="No services match your filters"
              description="Try a different search term, or clear team/tier filters."
              action={
                <button
                  onClick={() => {
                    setSearch("");
                    setTeam("");
                    setTier("");
                  }}
                  className="btn btn-secondary"
                >
                  Clear filters
                </button>
              }
            />
          )}

          {allServices.length === 0 && (
            <EmptyState
              icon={<IconServices />}
              title="No services registered yet"
              description="Scaffold your first service with the golden path — it will show up here automatically once catalog-info.yaml lands on main."
              action={
                <Link to="/create" className="btn btn-primary">
                  Create a service
                </Link>
              }
            />
          )}

          {filtered.length > 0 &&
            (view === "grid" ? (
              <section className="grid gap-4 md:grid-cols-2">
                {gridSorted.map((service, i) => (
                  <ServiceCard key={service.id} service={service} index={i} />
                ))}
              </section>
            ) : (
              <ServiceDataTable services={tableSorted} sortKey={tableSortKey} sortDir={tableSortDir} onSort={handleTableSort} />
            ))}
        </>
      )}
    </div>
  );
}

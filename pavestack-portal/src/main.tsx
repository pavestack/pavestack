import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import type { Catalog, CatalogService, CatalogCriteria, SortKey } from "./lib/catalog";
import {
  scoreColor,
  scoreDashOffset,
  scoreTier,
  filterServices,
  sortServices,
  computeStats,
} from "./lib/catalog";
import "./styles.css";

/* ────────────────────────────── Icons (inline SVG) ──────────────────────────── */

function IconSearch() {
  return (
    <svg className="w-4 h-4 text-pave-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>
  );
}

function IconCheck() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
  );
}

function IconX() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
  );
}

function IconExternalLink() {
  return (
    <svg className="w-3.5 h-3.5 inline-block ml-1 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
    </svg>
  );
}

function IconServices() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
    </svg>
  );
}

function IconScore() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
    </svg>
  );
}

function IconPassing() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  );
}

/* ────────────────────────────── Score Ring ──────────────────────────── */

const RING_RADIUS = 40;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;

function ScoreRing({ score, size = 80 }: { score: number; size?: number }) {
  const color = scoreColor(score);
  const offset = scoreDashOffset(score, RING_CIRCUMFERENCE);

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg viewBox="0 0 100 100" className="score-ring" width={size} height={size}>
        <circle cx="50" cy="50" r={RING_RADIUS} className="score-ring-track" />
        <circle
          cx="50"
          cy="50"
          r={RING_RADIUS}
          className="score-ring-fill animate-score-fill"
          style={{
            stroke: color,
            strokeDasharray: RING_CIRCUMFERENCE,
            strokeDashoffset: offset,
          }}
        />
      </svg>
      <span
        className="absolute text-lg font-bold tabular-nums"
        style={{ color }}
      >
        {score}
      </span>
    </div>
  );
}

/* ────────────────────────────── Stat Card ──────────────────────────── */

function StatCard({
  icon,
  label,
  value,
  subtext,
  delay,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subtext?: string;
  delay: number;
}) {
  return (
    <div className={`stat-card animate-slide-up`} style={{ animationDelay: `${delay}s` }}>
      <div className="flex items-center gap-2 text-pave-500 mb-2">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
      </div>
      <div className="text-2xl font-bold text-pave-300 tabular-nums">{value}</div>
      {subtext && <div className="text-xs text-pave-600 mt-1">{subtext}</div>}
    </div>
  );
}

/* ────────────────────────────── Criteria Row ──────────────────────────── */

function CriteriaRow({ item }: { item: CatalogCriteria }) {
  const passing = item.status === "passing";
  return (
    <div className="flex items-center justify-between gap-3 py-1.5">
      <div className="flex items-center gap-2 min-w-0">
        <span className={passing ? "text-pave-success" : "text-pave-danger"}>
          {passing ? <IconCheck /> : <IconX />}
        </span>
        <span className="text-sm text-pave-400 truncate capitalize">
          {item.key.replaceAll("_", " ")}
        </span>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <span className={`badge text-[10px] ${passing ? "badge-success" : "badge-danger"}`}>
          {item.status}
        </span>
        <span className="text-xs text-pave-600 tabular-nums w-6 text-right">{item.weight}%</span>
      </div>
    </div>
  );
}

/* ────────────────────────────── Environment Badge ──────────────────────────── */

function EnvBadge({ env, status, health }: { env: string; status: string; health: string }) {
  const synced = status === "synced" || status === "Synced";
  const healthy = health === "healthy" || health === "Healthy";
  const allGood = synced && healthy;

  return (
    <div
      className={`flex items-center gap-2 rounded-lg px-3 py-2 text-xs border ${
        allGood
          ? "bg-pave-success/5 border-pave-success/20"
          : "bg-pave-warning/5 border-pave-warning/20"
      }`}
    >
      <span
        className={`w-2 h-2 rounded-full ${
          allGood ? "bg-pave-success animate-pulse-slow" : "bg-pave-warning"
        }`}
      />
      <span className="font-semibold text-pave-300 uppercase">{env}</span>
      <span className={allGood ? "text-pave-success" : "text-pave-warning"}>
        {status} · {health}
      </span>
    </div>
  );
}

/* ────────────────────────────── Service Card ──────────────────────────── */

function ServiceCard({ service, index }: { service: CatalogService; index: number }) {
  const [expanded, setExpanded] = useState(false);
  const tier = scoreTier(service.scorecard.overall);

  return (
    <article
      id={`service-${service.id}`}
      className={`card p-6 animate-slide-up`}
      style={{ animationDelay: `${0.1 + index * 0.05}s` }}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h2 className="text-lg font-semibold text-pave-300 truncate">{service.name}</h2>
            <span className={`badge ${tier === "excellent" ? "badge-success" : tier === "good" ? "badge-accent" : tier === "warning" ? "badge-warning" : "badge-danger"}`}>
              {tier}
            </span>
          </div>
          <p className="text-sm text-pave-500 leading-relaxed">{service.description}</p>
        </div>
        <ScoreRing score={service.scorecard.overall} size={72} />
      </div>

      {/* Metadata Grid */}
      <div className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm mb-5">
        <div>
          <dt className="text-xs text-pave-600 uppercase tracking-wider mb-0.5">Owner</dt>
          <dd className="font-medium text-pave-400">{service.owner}</dd>
        </div>
        <div>
          <dt className="text-xs text-pave-600 uppercase tracking-wider mb-0.5">Lifecycle</dt>
          <dd className="font-medium text-pave-400 capitalize">{service.lifecycle}</dd>
        </div>
        <div className="col-span-2">
          <dt className="text-xs text-pave-600 uppercase tracking-wider mb-0.5">Repository</dt>
          <dd>
            <a
              className="text-pave-accent hover:text-pave-accent-hover transition-colors text-sm"
              href={service.repo}
              target="_blank"
              rel="noreferrer"
            >
              {service.repoPath}
              <IconExternalLink />
            </a>
          </dd>
        </div>
      </div>

      {/* Environments */}
      <div className="mb-5">
        <h3 className="text-xs font-medium uppercase text-pave-600 tracking-wider mb-2">Environments</h3>
        <div className="flex flex-wrap gap-2">
          {Object.entries(service.environments).map(([env, state]) => (
            <EnvBadge key={env} env={env} status={state.status} health={state.health} />
          ))}
        </div>
      </div>

      {/* Scorecard Toggle */}
      <div>
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-2 text-xs font-medium uppercase text-pave-500 hover:text-pave-accent transition-colors tracking-wider group"
          aria-expanded={expanded}
        >
          <svg
            className={`w-3 h-3 transition-transform duration-200 ${expanded ? "rotate-90" : ""}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
          </svg>
          Scorecard Details
          <span className="text-pave-600 font-normal normal-case tracking-normal">
            ({service.scorecard.criteria.filter((c) => c.status === "passing").length}/{service.scorecard.criteria.length} passing)
          </span>
        </button>
        {expanded && (
          <div className="mt-3 pt-3 border-t border-pave-800 animate-fade-in">
            {service.scorecard.criteria.map((item) => (
              <CriteriaRow key={item.key} item={item} />
            ))}
          </div>
        )}
      </div>
    </article>
  );
}

/* ────────────────────────────── Sort Selector ──────────────────────────── */

function SortSelect({ value, onChange }: { value: SortKey; onChange: (v: SortKey) => void }) {
  return (
    <select
      id="sort-select"
      value={value}
      onChange={(e) => onChange(e.target.value as SortKey)}
      className="rounded-lg border border-pave-700 bg-pave-850 px-3 py-2.5 text-sm text-pave-400 outline-none
                 transition-all duration-200 hover:border-pave-600 focus:border-pave-accent/50 focus:ring-1 focus:ring-pave-accent/20"
    >
      <option value="name">Sort by Name</option>
      <option value="score">Sort by Score</option>
      <option value="owner">Sort by Owner</option>
    </select>
  );
}

/* ────────────────────────────── Main App ──────────────────────────── */

function App() {
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

/* ────────────────────────────── Mount ──────────────────────────── */

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

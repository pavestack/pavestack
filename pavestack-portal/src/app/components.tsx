import React, { useState } from "react";
import type { CatalogService, CatalogCriteria, SortKey } from "../lib/catalog";
import {
  scoreColor,
  scoreDashOffset,
  scoreTier,
} from "../lib/catalog";
import {
  IconCheck,
  IconX,
  IconExternalLink,
} from "./icons";

/* ────────────────────────────── Score Ring ──────────────────────────── */

const RING_RADIUS = 40;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;

export function ScoreRing({ score, size = 80 }: { score: number; size?: number }) {
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

export function StatCard({
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

export function CriteriaRow({ item }: { item: CatalogCriteria }) {
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

export function EnvBadge({ env, status, health }: { env: string; status: string; health: string }) {
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

/* ────────────────────────────── Sort Selector ──────────────────────────── */

export function SortSelect({ value, onChange }: { value: SortKey; onChange: (v: SortKey) => void }) {
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

/* ────────────────────────────── Service Card ──────────────────────────── */

export function ServiceCard({ service, index }: { service: CatalogService; index: number }) {
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

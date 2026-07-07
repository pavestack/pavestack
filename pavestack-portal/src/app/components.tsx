import React, { useState } from "react";
import { Link } from "react-router-dom";
import type {
  CatalogService,
  CatalogCriteria,
  SortKey,
  Tier,
  PolicyCompliance,
  CostSummary,
  DeploymentHealth,
  ApiSummary,
} from "../lib/catalog";
import {
  scoreColor,
  scoreDashOffset,
  scoreTier,
  policyComplianceLabel,
  policyComplianceTier,
  costLabel,
  deploymentHealthLabel,
  deploymentHealthTier,
} from "../lib/catalog";
import { IconAlertTriangle, IconCheck, IconExternalLink, IconInbox, IconX } from "./icons";

/** Map a score/signal tier to the shared badge color classes. */
function tierBadgeClass(tier: "excellent" | "good" | "warning" | "critical" | "unknown"): string {
  switch (tier) {
    case "excellent":
      return "badge-success";
    case "good":
      return "badge-accent";
    case "warning":
      return "badge-warning";
    case "critical":
      return "badge-danger";
    default:
      return "badge-neutral";
  }
}

/* ────────────────────────────── Score Ring ──────────────────────────── */

const RING_RADIUS = 40;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;

export function ScoreRing({ score, size = 80 }: { score: number; size?: number }) {
  const color = scoreColor(score);
  const offset = scoreDashOffset(score, RING_CIRCUMFERENCE);

  return (
    <div
      className="relative inline-flex items-center justify-center"
      style={{ width: size, height: size }}
    >
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
      <span className="absolute text-lg font-bold tabular-nums" style={{ color }}>
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
  delay = 0,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subtext?: string;
  delay?: number;
}) {
  return (
    <div className="stat-card animate-slide-up" style={{ animationDelay: `${delay}s` }}>
      <div className="flex items-center gap-2 text-pave-text-muted mb-2">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
      </div>
      <div className="text-xl font-bold text-pave-text tabular-nums">{value}</div>
      {subtext && <div className="text-xs text-pave-text-muted mt-1">{subtext}</div>}
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
        <span className="text-sm text-pave-text-secondary truncate">{item.label}</span>
        {item.evidence && (
          <span className="text-xs text-pave-text-muted truncate font-mono hidden sm:inline">
            · {item.evidence}
          </span>
        )}
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <span className={`badge text-[10px] ${passing ? "badge-success" : "badge-danger"}`}>
          {item.status}
        </span>
        <span className="text-xs text-pave-text-muted tabular-nums w-8 text-right">
          {item.weight}%
        </span>
      </div>
    </div>
  );
}

/* ────────────────────────────── Environment Badge ──────────────────────────── */

export function EnvBadge({
  env,
  status,
  health,
  imageTag,
}: {
  env: string;
  status: string;
  health: string;
  imageTag?: string | null;
}) {
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
        className={`w-2 h-2 rounded-full shrink-0 ${allGood ? "bg-pave-success animate-pulse-slow" : "bg-pave-warning"}`}
      />
      <span className="font-semibold text-pave-text uppercase">{env}</span>
      <span className={allGood ? "text-pave-success" : "text-pave-warning"}>
        {status} · {health}
      </span>
      {imageTag && <span className="font-mono tabular-nums text-pave-text-muted">{imageTag}</span>}
    </div>
  );
}

/* ────────────────────────────── Platform Signals ──────────────────────────── */

/** One environment's row of report-derived scorecard signals: policy, cost, deployment health. */
export function PlatformSignalsRow({
  env,
  policyCompliance,
  costPerMonth,
  deploymentHealth,
}: {
  env: string;
  policyCompliance: PolicyCompliance;
  costPerMonth: CostSummary;
  deploymentHealth: DeploymentHealth;
}) {
  return (
    <div className="flex items-center justify-between gap-3 py-1.5">
      <span className="text-xs font-semibold uppercase text-pave-600 tracking-wider shrink-0">
        {env}
      </span>
      <div className="flex flex-wrap items-center justify-end gap-1.5">
        <span
          className={`badge text-[10px] ${tierBadgeClass(policyComplianceTier(policyCompliance))}`}
        >
          Policy {policyComplianceLabel(policyCompliance)}
        </span>
        <span className="badge badge-neutral text-[10px]">Cost {costLabel(costPerMonth)}</span>
        <span
          className={`badge text-[10px] ${tierBadgeClass(deploymentHealthTier(deploymentHealth))}`}
        >
          {deploymentHealthLabel(deploymentHealth)}
        </span>
      </div>
    </div>
  );
}

/* ────────────────────────────── API Section ──────────────────────────── */

/** Renders the endpoint table parsed from <service>/openapi.yaml. Renders nothing if absent. */
export function ApiSection({ api }: { api: ApiSummary | undefined }) {
  if (!api) return null;

  return (
    <div className="mt-5 pt-5 border-t border-pave-800">
      <div className="flex items-center justify-between gap-3 mb-2">
        <h3 className="text-xs font-medium uppercase text-pave-600 tracking-wider">API</h3>
        <span className="text-xs text-pave-600 font-mono truncate">
          {api.title}
          {api.version ? ` v${api.version}` : ""}
        </span>
      </div>
      {api.endpoints.length === 0 ? (
        <p className="text-xs text-pave-600">No endpoints documented.</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-pave-800">
          <table className="w-full text-xs">
            <thead className="bg-pave-850 text-pave-600 uppercase tracking-wider">
              <tr>
                <th className="text-left px-3 py-2 font-medium">Method</th>
                <th className="text-left px-3 py-2 font-medium">Path</th>
                <th className="text-left px-3 py-2 font-medium">Summary</th>
              </tr>
            </thead>
            <tbody>
              {api.endpoints.map((ep, i) => (
                <tr key={`${ep.method}-${ep.path}-${i}`} className="border-t border-pave-800">
                  <td className="px-3 py-2 font-mono text-pave-accent">{ep.method}</td>
                  <td className="px-3 py-2 font-mono text-pave-400">{ep.path}</td>
                  <td className="px-3 py-2 text-pave-500">{ep.summary || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

/* ────────────────────────────── Sort Selector ──────────────────────────── */

export function SortSelect({
  value,
  onChange,
}: {
  value: SortKey;
  onChange: (v: SortKey) => void;
}) {
  return (
    <select
      id="sort-select"
      aria-label="Sort services"
      value={value}
      onChange={(e) => onChange(e.target.value as SortKey)}
      className="rounded-lg border border-pave-border bg-pave-surface px-3 py-2 text-sm text-pave-text-secondary outline-none
                 transition-colors duration-150 hover:border-pave-border-strong focus:border-pave-accent/50 focus:ring-1 focus:ring-pave-accent/20"
    >
      <option value="name">Sort by Name</option>
      <option value="score">Sort by Score</option>
      <option value="owner">Sort by Owner</option>
    </select>
  );
}

/* ────────────────────────────── Tier / Runtime / CreatedVia badges ──────────────────────────── */

export function TierBadge({ tier }: { tier: Tier }) {
  if (!tier) return <span className="badge badge-neutral">no tier set</span>;
  const style =
    tier === "tier-1" ? "badge-danger" : tier === "tier-2" ? "badge-accent" : "badge-neutral";
  return <span className={`badge ${style}`}>{tier}</span>;
}

export function CreatedViaBadge({ createdVia }: { createdVia: CatalogService["createdVia"] }) {
  return createdVia === "pave-cli" ? (
    <span className="badge badge-success">pave-cli</span>
  ) : (
    <span className="badge badge-neutral">manual</span>
  );
}

/* ────────────────────────────── Service Card ──────────────────────────── */

export function ServiceCard({ service, index = 0 }: { service: CatalogService; index?: number }) {
  const [expanded, setExpanded] = useState(false);
  const tier = scoreTier(service.scorecard.overallScore);

  return (
    <article
      id={`service-${service.id}`}
      className="card p-5 animate-slide-up"
      style={{ animationDelay: `${Math.min(index * 0.03, 0.3)}s` }}
    >
      <div className="flex items-start justify-between gap-4 mb-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1 flex-wrap">
            <Link
              to={`/services/${service.name}`}
              className="text-base font-semibold text-pave-text truncate hover:text-pave-accent"
            >
              {service.name}
            </Link>
            <span
              className={`badge ${
                tier === "excellent"
                  ? "badge-success"
                  : tier === "good"
                    ? "badge-accent"
                    : tier === "warning"
                      ? "badge-warning"
                      : "badge-danger"
              }`}
            >
              {tier}
            </span>
            <CreatedViaBadge createdVia={service.createdVia} />
          </div>
          <p className="text-sm text-pave-text-secondary leading-relaxed">{service.description}</p>
        </div>
        <ScoreRing score={service.scorecard.overallScore} size={64} />
      </div>

      <div className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm mb-4">
        <div>
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-0.5">Owner</dt>
          <dd className="font-medium text-pave-text-secondary">{service.owner}</dd>
        </div>
        <div>
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-0.5">
            Lifecycle
          </dt>
          <dd className="font-medium text-pave-text-secondary capitalize">{service.lifecycle}</dd>
        </div>
        <div className="col-span-2">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-0.5">
            Repository
          </dt>
          <dd>
            <a
              className="text-pave-accent hover:text-pave-accent-hover transition-colors text-sm"
              href={service.repoUrl}
              target="_blank"
              rel="noreferrer"
            >
              {service.repoPath}
              <IconExternalLink />
            </a>
          </dd>
        </div>
      </div>

      <div className="mb-4">
        <h3 className="text-xs font-medium uppercase text-pave-text-muted tracking-wider mb-2">
          Environments
        </h3>
        <div className="flex flex-wrap gap-2">
          {Object.entries(service.environments).map(([env, state]) => (
            <EnvBadge
              key={env}
              env={env}
              status={state.status}
              health={state.health}
              imageTag={state.imageTag}
            />
          ))}
        </div>
      </div>

      <div>
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-2 text-xs font-medium uppercase text-pave-text-secondary hover:text-pave-accent transition-colors tracking-wider"
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
          Scorecard details
          <span className="text-pave-text-muted font-normal normal-case tracking-normal">
            ({service.scorecard.criteria.filter((c) => c.status === "passing").length}/
            {service.scorecard.criteria.length} passing)
          </span>
        </button>
        {expanded && (
          <div className="mt-3 pt-3 border-t border-pave-border animate-fade-in">
            {service.scorecard.criteria.map((item) => (
              <CriteriaRow key={item.key} item={item} />
            ))}
            {(service.policyCompliance || service.costPerMonth || service.deploymentHealth) && (
              <div className="mt-3 pt-3 border-t border-pave-800">
                <h3 className="text-xs font-medium uppercase text-pave-600 tracking-wider mb-1">
                  Platform Signals
                </h3>
                {Object.keys(service.environments).map((env) => (
                  <PlatformSignalsRow
                    key={env}
                    env={env}
                    policyCompliance={service.policyCompliance?.[env] ?? null}
                    costPerMonth={service.costPerMonth?.[env] ?? null}
                    deploymentHealth={service.deploymentHealth?.[env] ?? null}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      <ApiSection api={service.api} />
    </article>
  );
}

/* ────────────────────────────── Loading / empty / error states ──────────────────────────── */

export function Skeleton({ className = "" }: { className?: string }) {
  return <div className={`skeleton ${className}`} aria-hidden="true" />;
}

export function SkeletonGrid({ rows = 3 }: { rows?: number }) {
  return (
    <div className="grid gap-4 md:grid-cols-2" role="status" aria-label="Loading">
      {Array.from({ length: rows }, (_, i) => (
        <div key={i} className="card p-5 space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-5 w-40" />
            <Skeleton className="h-12 w-12 rounded-full" />
          </div>
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-2/3" />
        </div>
      ))}
    </div>
  );
}

export function EmptyState({
  icon,
  title,
  description,
  action,
}: {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="text-center py-16 px-6 animate-fade-in">
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-pave-surface-hover text-pave-text-muted mb-4">
        {icon ?? <IconInbox />}
      </div>
      <p className="text-pave-text text-base font-medium mb-1">{title}</p>
      {description && (
        <p className="text-pave-text-muted text-sm max-w-md mx-auto mb-4">{description}</p>
      )}
      {action}
    </div>
  );
}

export function InlineError({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-3 rounded-lg border border-pave-danger/30 bg-pave-danger/5 px-4 py-3 text-sm text-pave-danger"
    >
      <IconAlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
      <div className="flex-1">
        <p>{message}</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="mt-1 font-medium underline underline-offset-2 hover:no-underline"
          >
            Try again
          </button>
        )}
      </div>
    </div>
  );
}

export function DemoDataNote({ count }: { count: number }) {
  return (
    <div className="flex items-center gap-2 rounded-lg border border-pave-info/25 bg-pave-info/5 px-3 py-2 text-xs text-pave-info">
      <IconAlertTriangle className="w-3.5 h-3.5 shrink-0" />
      <span>
        Showing <strong className="tabular-nums">{count}</strong> synthetic demo rows appended for
        scale/virtualization illustration — not real services.
      </span>
    </div>
  );
}

export function SimulatedNote({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2 rounded-lg border border-pave-warning/25 bg-pave-warning/5 px-3 py-2 text-xs text-pave-warning">
      <IconAlertTriangle className="w-3.5 h-3.5 shrink-0 mt-0.5" />
      <span>{children}</span>
    </div>
  );
}

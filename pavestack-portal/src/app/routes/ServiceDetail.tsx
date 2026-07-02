import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useCatalog } from "../CatalogContext";
import { CreatedViaBadge, CriteriaRow, EmptyState, InlineError, ScoreRing, SimulatedNote, Skeleton, TierBadge } from "../components";
import { getCostEstimate, type CostEstimate } from "../../lib/api";
import { IconExternalLink, IconServices } from "../icons";

export function ServiceDetail() {
  const { name } = useParams<{ name: string }>();
  const { catalog, loading } = useCatalog();

  const service = catalog?.services.find((s) => s.name === name);

  const [cost, setCost] = useState<CostEstimate | null>(null);
  const [costError, setCostError] = useState<string | null>(null);
  const [costLoading, setCostLoading] = useState(false);

  useEffect(() => {
    if (!service || !service.tier || !service.exposure) return;
    setCostLoading(true);
    setCostError(null);
    getCostEstimate({ tier: service.tier, exposure: service.exposure, database: Boolean(service.database) })
      .then((data) => setCost(data))
      .catch((err) => setCostError(err instanceof Error ? err.message : "Could not reach pave-api"))
      .finally(() => setCostLoading(false));
  }, [service?.name, service?.tier, service?.exposure, service?.database]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!service) {
    return (
      <EmptyState
        icon={<IconServices />}
        title="Service not found"
        description={`No catalog entry named "${name}". It may not exist, or catalog.json hasn't been regenerated yet.`}
        action={
          <Link to="/" className="btn btn-primary">
            Back to catalog
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <Link to="/" className="text-xs text-pave-text-muted hover:text-pave-accent">
          ← Back to catalog
        </Link>
        <div className="flex flex-wrap items-center gap-2 mt-2">
          <h1 className="text-xl font-semibold text-pave-text">{service.name}</h1>
          <TierBadge tier={service.tier} />
          <CreatedViaBadge createdVia={service.createdVia} />
          <span className="badge badge-neutral capitalize">{service.lifecycle}</span>
        </div>
        <p className="text-sm text-pave-text-secondary mt-1 max-w-2xl">{service.description}</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">Team</dt>
          <dd className="text-sm font-medium text-pave-text">{service.team}</dd>
        </div>
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">Runtime</dt>
          <dd className="text-sm font-medium text-pave-text">{service.runtime ?? "not set"}</dd>
        </div>
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">Exposure</dt>
          <dd className="text-sm font-medium text-pave-text">{service.exposure ?? "not set"}</dd>
        </div>
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">Repository</dt>
          <dd className="text-sm">
            <a href={service.repoUrl} target="_blank" rel="noreferrer" className="text-pave-accent hover:text-pave-accent-hover">
              {service.repoPath}
              <IconExternalLink />
            </a>
          </dd>
        </div>
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">Managed database</dt>
          <dd className="text-sm font-medium text-pave-text">{service.database === null ? "unknown" : service.database ? "Yes" : "No"}</dd>
        </div>
        <div className="card p-4">
          <dt className="text-xs text-pave-text-muted uppercase tracking-wider mb-1">System</dt>
          <dd className="text-sm font-medium text-pave-text">{service.system}</dd>
        </div>
      </div>

      {/* Scorecard */}
      <section className="card p-5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-base font-semibold text-pave-text">Scorecard</h2>
          <Link to="/scorecards" className="text-xs text-pave-accent hover:underline">
            View leaderboard →
          </Link>
        </div>
        <div className="flex flex-col sm:flex-row gap-6">
          <div className="flex justify-center sm:justify-start">
            <ScoreRing score={service.scorecard.overallScore} size={96} />
          </div>
          <div className="flex-1 divide-y divide-pave-border">
            {service.scorecard.criteria.map((item) => (
              <CriteriaRow key={item.key} item={item} />
            ))}
          </div>
        </div>
      </section>

      {/* Environments / Argo CD */}
      <section className="card p-5">
        <h2 className="text-base font-semibold text-pave-text mb-1">Environments</h2>
        <div className="mb-3">
          <SimulatedNote>
            Sync status and health below are illustrative sample data pending a live Argo CD integration. Image tags are real —
            read directly from the committed Helm values in <code className="font-mono">platform-config/tenants/{service.name}</code>.
          </SimulatedNote>
        </div>
        <div className="grid gap-3 sm:grid-cols-2">
          {Object.entries(service.environments).map(([env, state]) => (
            <div key={env} className="rounded-lg border border-pave-border p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-semibold uppercase text-pave-text">{env}</span>
                <span className={`badge ${state.status === "synced" ? "badge-success" : "badge-warning"}`}>{state.status}</span>
              </div>
              <dl className="text-sm space-y-1">
                <div className="flex justify-between">
                  <dt className="text-pave-text-muted">Health</dt>
                  <dd className="text-pave-text-secondary capitalize">{state.health}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-pave-text-muted">Image tag</dt>
                  <dd className="font-mono tabular-nums text-pave-text-secondary">{state.imageTag ?? "unknown"}</dd>
                </div>
              </dl>
            </div>
          ))}
        </div>
      </section>

      {/* Deployment history */}
      <section className="card p-5">
        <h2 className="text-base font-semibold text-pave-text mb-2">Deployment history</h2>
        <SimulatedNote>
          History not yet wired to a real deploy log — showing current state only. A future iteration will read this from CI
          release events or Argo CD application history.
        </SimulatedNote>
      </section>

      {/* Resource usage */}
      <section className="card p-5">
        <h2 className="text-base font-semibold text-pave-text mb-2">Resource usage</h2>

        {Object.values(service.environments).some((e) => e.resources) && (
          <div className="grid gap-3 sm:grid-cols-2 mb-4">
            {Object.entries(service.environments).map(([env, state]) =>
              state.resources ? (
                <div key={env} className="rounded-lg border border-pave-border p-4">
                  <p className="text-xs font-semibold uppercase text-pave-text-muted mb-2">
                    {env} · configured allocation
                  </p>
                  <dl className="text-sm space-y-1">
                    <div className="flex justify-between">
                      <dt className="text-pave-text-muted">CPU request / limit</dt>
                      <dd className="font-mono tabular-nums text-pave-text-secondary">
                        {state.resources.requests?.cpu ?? "?"} / {state.resources.limits?.cpu ?? "?"}
                      </dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-pave-text-muted">Memory request / limit</dt>
                      <dd className="font-mono tabular-nums text-pave-text-secondary">
                        {state.resources.requests?.memory ?? "?"} / {state.resources.limits?.memory ?? "?"}
                      </dd>
                    </div>
                  </dl>
                </div>
              ) : null
            )}
          </div>
        )}

        <EmptyState
          title="No live usage metrics connected"
          description="Actual CPU/memory consumption and request-rate charts will appear here once Prometheus/Grafana wiring lands. The configured allocation above is real (from committed Helm values) — no usage numbers are fabricated in the meantime."
        />
      </section>

      {/* Cost estimate */}
      <section className="card p-5">
        <h2 className="text-base font-semibold text-pave-text mb-3">Cost estimate</h2>
        {!service.tier || !service.exposure ? (
          <p className="text-sm text-pave-text-muted">
            Tier and exposure aren't set for this service, so a cost estimate can't be computed. Services scaffolded through
            the Create Service wizard will have this set automatically.
          </p>
        ) : costLoading ? (
          <Skeleton className="h-16 w-full" />
        ) : costError ? (
          <InlineError message={`Backend unreachable — start pave-api locally to see a live cost estimate. (${costError})`} />
        ) : cost ? (
          <div>
            <p className="text-2xl font-bold tabular-nums text-pave-text">
              ${cost.monthlyUsdLow}–${cost.monthlyUsdHigh}
              <span className="text-sm font-normal text-pave-text-muted ml-1">/mo estimated</span>
            </p>
            <ul className="mt-3 text-sm divide-y divide-pave-border">
              {cost.breakdown.map((b) => (
                <li key={b.item} className="flex justify-between py-1.5">
                  <span className="text-pave-text-secondary">{b.item}</span>
                  <span className="tabular-nums text-pave-text">${b.monthlyUsd}</span>
                </li>
              ))}
            </ul>
            <p className="text-xs text-pave-text-muted mt-3">{cost.disclaimer}</p>
          </div>
        ) : null}
      </section>
    </div>
  );
}

import React, { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useCatalog } from "../CatalogContext";
import { CreatedViaBadge, EmptyState, InlineError, ScoreRing, SkeletonGrid, TierBadge } from "../components";
import { IconTrophy } from "../icons";

export function Scorecards() {
  const { catalog, loading, error, reload } = useCatalog();
  const [sortAsc, setSortAsc] = useState(false);

  const ranked = useMemo(() => {
    const services = catalog?.services ?? [];
    return [...services].sort((a, b) =>
      sortAsc ? a.scorecard.overallScore - b.scorecard.overallScore : b.scorecard.overallScore - a.scorecard.overallScore
    );
  }, [catalog, sortAsc]);

  const topPerformers = ranked.filter((s) => s.scorecard.overallScore === 100);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-pave-text">Top performing services</h1>
        <p className="text-sm text-pave-text-muted mt-0.5">
          Scorecard compliance across the catalog — a shared bar the whole platform is climbing together.
        </p>
      </div>

      {loading && <SkeletonGrid rows={4} />}
      {!loading && error && <InlineError message="Failed to load catalog.json." onRetry={reload} />}

      {!loading && catalog && catalog.services.length === 0 && (
        <EmptyState icon={<IconTrophy />} title="No services scored yet" description="Scorecards populate automatically as services register scorecard.yaml." />
      )}

      {!loading && catalog && catalog.services.length > 0 && (
        <>
          {topPerformers.length > 0 && (
            <div className="mb-6 rounded-lg border border-pave-success/25 bg-pave-success/5 px-4 py-3 flex items-center gap-3">
              <IconTrophy className="w-5 h-5 text-pave-success shrink-0" />
              <p className="text-sm text-pave-success">
                <strong>{topPerformers.length}</strong> service{topPerformers.length === 1 ? "" : "s"} at a perfect 100 —{" "}
                {topPerformers.map((s) => s.name).join(", ")}.
              </p>
            </div>
          )}

          <div className="card overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-pave-border bg-pave-elevated text-left">
                  <th className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-pave-text-muted">Rank</th>
                  <th className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-pave-text-muted">Service</th>
                  <th className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-pave-text-muted">Team</th>
                  <th className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-pave-text-muted">Tier</th>
                  <th className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-pave-text-muted">Origin</th>
                  <th className="px-4 py-2.5">
                    <button
                      onClick={() => setSortAsc((v) => !v)}
                      className="th-sortable"
                      aria-sort={sortAsc ? "ascending" : "descending"}
                    >
                      Score {sortAsc ? "↑" : "↓"}
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                {ranked.map((service, i) => (
                  <tr key={service.id} className="border-b border-pave-border last:border-b-0 hover:bg-pave-surface-hover">
                    <td className="px-4 py-3 tabular-nums text-pave-text-muted">
                      {i === 0 && !sortAsc ? <span title="Top performer">🏆</span> : `#${i + 1}`}
                    </td>
                    <td className="px-4 py-3">
                      <Link to={`/services/${service.name}`} className="font-medium text-pave-text hover:text-pave-accent">
                        {service.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-pave-text-secondary">{service.team}</td>
                    <td className="px-4 py-3">
                      <TierBadge tier={service.tier} />
                    </td>
                    <td className="px-4 py-3">
                      <CreatedViaBadge createdVia={service.createdVia} />
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <ScoreRing score={service.scorecard.overallScore} size={32} />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}

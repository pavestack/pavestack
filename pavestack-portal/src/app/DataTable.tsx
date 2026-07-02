import React, { useRef } from "react";
import { Link } from "react-router-dom";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { CatalogService } from "../lib/catalog";
import { scoreTier } from "../lib/catalog";
import { CreatedViaBadge, TierBadge } from "./components";
import { IconChevronDown, IconChevronUp, IconUpDown } from "./icons";

export type TableSortKey = "name" | "team" | "lifecycle" | "tier" | "score" | "createdVia";
export type SortDir = "asc" | "desc";

const ROW_HEIGHT = 44;

function scoreBadgeClass(score: number) {
  const tier = scoreTier(score);
  if (tier === "excellent") return "text-pave-success";
  if (tier === "good") return "text-pave-info";
  if (tier === "warning") return "text-pave-warning";
  return "text-pave-danger";
}

const COLUMNS: { key: TableSortKey; label: string; numeric?: boolean; width: string }[] = [
  { key: "name", label: "Service", width: "minmax(180px, 2fr)" },
  { key: "team", label: "Team", width: "minmax(120px, 1fr)" },
  { key: "lifecycle", label: "Lifecycle", width: "120px" },
  { key: "tier", label: "Tier", width: "110px" },
  { key: "createdVia", label: "Created via", width: "120px" },
  { key: "score", label: "Score", numeric: true, width: "90px" },
];

function gridTemplate() {
  return COLUMNS.map((c) => c.width).join(" ");
}

// Sum of column minimums — used so the table scrolls horizontally as a unit
// on narrow viewports instead of clipping trailing columns.
const MIN_TABLE_WIDTH = 760;

/**
 * Virtualized, sortable data table. Renders only visible rows (windowing via
 * @tanstack/react-virtual) so 1000+ rows stay smooth without a real DOM row
 * per record.
 */
export function ServiceDataTable({
  services,
  sortKey,
  sortDir,
  onSort,
}: {
  services: CatalogService[];
  sortKey: TableSortKey;
  sortDir: SortDir;
  onSort: (key: TableSortKey) => void;
}) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: services.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 12,
  });

  const virtualRows = virtualizer.getVirtualItems();

  return (
    <div className="card overflow-x-auto" role="table" aria-label="Services" aria-rowcount={services.length + 1}>
      <div style={{ minWidth: MIN_TABLE_WIDTH }}>
        {/* Header row */}
        <div
          role="row"
          className="grid items-center gap-2 border-b border-pave-border bg-pave-elevated px-4 py-2.5 sticky top-0 z-10"
          style={{ gridTemplateColumns: gridTemplate() }}
        >
          {COLUMNS.map((col) => {
            const active = sortKey === col.key;
            return (
              <button
                key={col.key}
                role="columnheader"
                aria-sort={active ? (sortDir === "asc" ? "ascending" : "descending") : "none"}
                onClick={() => onSort(col.key)}
                className={`th-sortable flex items-center gap-1 ${col.numeric ? "justify-end" : ""}`}
              >
                {col.label}
                {active ? sortDir === "asc" ? <IconChevronUp /> : <IconChevronDown /> : <IconUpDown />}
              </button>
            );
          })}
        </div>

        {/* Virtualized body */}
        <div ref={parentRef} className="overflow-y-auto" style={{ maxHeight: "min(70vh, 640px)" }}>
          <div style={{ height: virtualizer.getTotalSize(), position: "relative" }}>
            {virtualRows.map((virtualRow) => {
              const service = services[virtualRow.index];
              return (
                <Link
                  key={service.id}
                  to={`/services/${service.name}`}
                  role="row"
                  aria-rowindex={virtualRow.index + 2}
                  className="grid items-center gap-2 px-4 border-b border-pave-border last:border-b-0 hover:bg-pave-surface-hover transition-colors duration-100 absolute left-0 right-0"
                  style={{
                    gridTemplateColumns: gridTemplate(),
                    height: ROW_HEIGHT,
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                >
                  <span role="cell" className="truncate text-sm font-medium text-pave-text flex items-center gap-2">
                    {service.name}
                    {service.isDemo && <span className="badge badge-neutral text-[9px] shrink-0">demo</span>}
                  </span>
                  <span role="cell" className="truncate text-sm text-pave-text-secondary">
                    {service.team}
                  </span>
                  <span role="cell" className="truncate text-sm text-pave-text-secondary capitalize">
                    {service.lifecycle}
                  </span>
                  <span role="cell">
                    <TierBadge tier={service.tier} />
                  </span>
                  <span role="cell">
                    <CreatedViaBadge createdVia={service.createdVia} />
                  </span>
                  <span role="cell" className={`text-sm font-semibold tabular-nums text-right ${scoreBadgeClass(service.scorecard.overallScore)}`}>
                    {service.scorecard.overallScore}
                  </span>
                </Link>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

export function sortServicesForTable(services: CatalogService[], key: TableSortKey, dir: SortDir): CatalogService[] {
  const sorted = [...services].sort((a, b) => {
    switch (key) {
      case "team":
        return a.team.localeCompare(b.team);
      case "lifecycle":
        return a.lifecycle.localeCompare(b.lifecycle);
      case "tier":
        return (a.tier ?? "").localeCompare(b.tier ?? "");
      case "createdVia":
        return a.createdVia.localeCompare(b.createdVia);
      case "score":
        return a.scorecard.overallScore - b.scorecard.overallScore;
      default:
        return a.name.localeCompare(b.name);
    }
  });
  return dir === "asc" ? sorted : sorted.reverse();
}

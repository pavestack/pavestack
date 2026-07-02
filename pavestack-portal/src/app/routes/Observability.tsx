import React, { useMemo, useState } from "react";
import { useCatalog } from "../CatalogContext";
import { generateSampleMetrics } from "../../lib/observability";
import { EmptyState, SimulatedNote, SkeletonGrid } from "../components";
import { Sparkline } from "../Sparkline";
import { IconActivity } from "../icons";

function MetricCard({ label, value, unit, series, color, markers }: { label: string; value: number; unit: string; series: number[]; color: string; markers?: number[] }) {
  return (
    <div className="card p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium uppercase tracking-wider text-pave-text-muted">{label}</span>
      </div>
      <p className="text-xl font-bold tabular-nums text-pave-text mb-2">
        {value}
        <span className="text-sm font-normal text-pave-text-muted ml-1">{unit}</span>
      </p>
      <Sparkline values={series} color={color} markers={markers} width={200} height={40} />
    </div>
  );
}

export function Observability() {
  const { catalog, loading } = useCatalog();
  const [selected, setSelected] = useState<string>("");

  const services = catalog?.services ?? [];
  const activeName = selected || services[0]?.name || "";
  const metrics = useMemo(() => (activeName ? generateSampleMetrics(activeName) : null), [activeName]);

  return (
    <div>
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-pave-text">Observability</h1>
        <p className="text-sm text-pave-text-muted mt-0.5">Golden-signals snapshot per service.</p>
      </div>

      <div className="mb-4">
        <SimulatedNote>
          No live metrics backend is wired up yet — the charts below are illustrative sample data generated client-side
          (deterministic per service), standing in for a future Prometheus/Grafana integration.
        </SimulatedNote>
      </div>

      {loading && <SkeletonGrid rows={2} />}

      {!loading && services.length === 0 && (
        <EmptyState icon={<IconActivity />} title="No services to observe" description="Metrics views populate once services are registered in the catalog." />
      )}

      {!loading && services.length > 0 && metrics && (
        <>
          <div className="mb-4">
            <label htmlFor="observability-service" className="block text-xs font-medium uppercase tracking-wider text-pave-text-muted mb-1">
              Service
            </label>
            <select
              id="observability-service"
              value={activeName}
              onChange={(e) => setSelected(e.target.value)}
              className="rounded-lg border border-pave-border bg-pave-surface px-3 py-2 text-sm text-pave-text-secondary outline-none hover:border-pave-border-strong"
            >
              {services.map((s) => (
                <option key={s.id} value={s.name}>
                  {s.name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-6">
            <MetricCard label="Request rate" value={metrics.currentRequestRate} unit="req/s" series={metrics.requestRate} color="var(--accent)" />
            <MetricCard
              label="Error rate"
              value={metrics.currentErrorRate}
              unit="%"
              series={metrics.errorRate}
              color="var(--danger)"
              markers={metrics.errorSpikeMarkers}
            />
            <MetricCard label="Latency p99" value={metrics.currentLatencyP99} unit="ms" series={metrics.latencyP99} color="var(--warning)" />
            <div className="card p-4">
              <span className="text-xs font-medium uppercase tracking-wider text-pave-text-muted">SLO burn rate</span>
              <p className={`text-xl font-bold tabular-nums mt-2 ${metrics.sloBurnRate > 1 ? "text-pave-danger" : "text-pave-success"}`}>
                {metrics.sloBurnRate}x
              </p>
              <p className="text-xs text-pave-text-muted mt-1">{metrics.sloBurnRate > 1 ? "Burning faster than budget allows" : "Within error budget"}</p>
            </div>
          </div>

          <section className="card p-5">
            <h2 className="text-sm font-semibold text-pave-text mb-2">Recent deploys vs. error-rate spikes</h2>
            {metrics.deployMarkers.length === 0 ? (
              <p className="text-sm text-pave-text-muted">No sample deploy events in this window.</p>
            ) : (
              <ul className="text-sm space-y-1.5">
                {metrics.deployMarkers.map((idx) => {
                  const spiked = metrics.errorSpikeMarkers.includes(idx);
                  return (
                    <li key={idx} className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-pave-accent shrink-0" />
                      <span className="text-pave-text-secondary tabular-nums">t-{POINTS_LABEL(idx)}</span>
                      <span className="text-pave-text-muted">deploy</span>
                      {spiked && <span className="badge badge-danger">correlated error-rate spike</span>}
                    </li>
                  );
                })}
              </ul>
            )}
            <p className="text-xs text-pave-text-muted mt-3">Sample annotated timeline — deploy/error correlation will be computed from real CI + metrics events once wired.</p>
          </section>
        </>
      )}
    </div>
  );
}

function POINTS_LABEL(idx: number): string {
  return `${23 - idx}h`;
}

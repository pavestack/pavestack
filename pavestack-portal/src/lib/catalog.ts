export type CatalogCriteria = {
  key: string;
  label: string;
  status: string;
  weight: number;
  evidence?: string | null;
};

export type ResourceBlock = { cpu: string | null; memory: string | null } | null;

export type EnvironmentState = {
  status: string;
  health: string;
  /** Current deployed image tag, read from the tenant's committed Helm values. Real when present. */
  imageTag?: string | null;
  /** Configured CPU/memory requests+limits from committed Helm values — real allocation, not live usage. */
  resources?: { requests: ResourceBlock; limits: ResourceBlock } | null;
};

/** Kyverno PolicyReport summary for one environment, or null when unknown (see reports/README.md). */
export type PolicyCompliance = {
  pass: number;
  warn: number;
  fail: number;
  passPercent: number;
} | null;

/** OpenCost monthly allocation for one environment, or null when unknown. */
export type CostSummary = {
  amount: number;
  currency: string;
} | null;

/** Argo CD app sync/health for one environment, or null when unknown. */
export type DeploymentHealth = {
  syncStatus: string;
  health: string;
  lastSyncAt: string | null;
} | null;

export type ApiEndpoint = {
  method: string;
  path: string;
  summary: string;
};

/** Parsed from <service>/openapi.yaml at build time, or null when the service has none. */
export type ApiSummary = {
  title: string;
  version: string;
  endpoints: ApiEndpoint[];
} | null;
export type Tier = "tier-1" | "tier-2" | "tier-3" | null;
export type CreatedVia = "pave-cli" | "manual";

export type CatalogService = {
  id: string;
  name: string;
  description: string;
  owner: string;
  team: string;
  system: string;
  repoUrl: string;
  repoPath: string;
  lifecycle: string;
  tier: Tier;
  runtime: string | null;
  exposure: "internal" | "public" | null;
  /** Whether the tenant is provisioned with a managed database, read from tenant.yaml. Null when unknown. */
  database: boolean | null;
  createdVia: CreatedVia;
  environments: Record<string, EnvironmentState>;
  scorecard: {
    overallScore: number;
    criteria: CatalogCriteria[];
  };
  /** Per-environment report-derived signals. Optional for backward compatibility with older catalogs. */
  policyCompliance?: Record<string, PolicyCompliance>;
  costPerMonth?: Record<string, CostSummary>;
  deploymentHealth?: Record<string, DeploymentHealth>;
  api?: ApiSummary;
  /** Present + true only for synthetic rows generated client-side to illustrate scale. Never set for real catalog data. */
  isDemo?: boolean;
};

export type Catalog = {
  generatedAt: string;
  services: CatalogService[];
};

export async function loadCatalog(): Promise<Catalog> {
  const response = await fetch("/catalog.json");
  if (!response.ok) {
    throw new Error("Failed to load catalog");
  }
  return response.json();
}

/** Classify a score into a tier for color coding */
export function scoreTier(score: number): "excellent" | "good" | "warning" | "critical" {
  if (score >= 90) return "excellent";
  if (score >= 70) return "good";
  if (score >= 50) return "warning";
  return "critical";
}

/**
 * Get the semantic color token for a score. Returns a `var(--token)`
 * reference (rather than a hardcoded hex) so score colors stay correct
 * across the light/dark theme toggle without any recomputation.
 */
export function scoreColor(score: number): string {
  if (score >= 90) return "var(--success)";
  if (score >= 70) return "var(--info)";
  if (score >= 50) return "var(--warning)";
  return "var(--danger)";
}

/** Calculate the SVG stroke-dashoffset for a circular progress ring */
export function scoreDashOffset(score: number, circumference: number): number {
  return circumference - (score / 100) * circumference;
}

/** Filter services by search query (name, owner/team, description) */
export function filterServices(services: CatalogService[], query: string): CatalogService[] {
  if (!query.trim()) return services;
  const lower = query.toLowerCase();
  return services.filter(
    (s) =>
      s.name.toLowerCase().includes(lower) ||
      s.owner.toLowerCase().includes(lower) ||
      s.team.toLowerCase().includes(lower) ||
      s.description.toLowerCase().includes(lower)
  );
}

export function filterByTeam(services: CatalogService[], team: string): CatalogService[] {
  if (!team) return services;
  return services.filter((s) => s.team === team);
}

export function filterByTier(services: CatalogService[], tier: string): CatalogService[] {
  if (!tier) return services;
  return services.filter((s) => s.tier === tier);
}

/** Sort services by various criteria */
export type SortKey = "name" | "score" | "owner";

export function sortServices(services: CatalogService[], key: SortKey): CatalogService[] {
  return [...services].sort((a, b) => {
    switch (key) {
      case "score":
        return b.scorecard.overallScore - a.scorecard.overallScore;
      case "owner":
        return a.owner.localeCompare(b.owner);
      default:
        return a.name.localeCompare(b.name);
    }
  });
}

/** Format a policy compliance signal for display, e.g. "92%" or "Unknown". */
export function policyComplianceLabel(compliance: PolicyCompliance): string {
  if (!compliance) return "Unknown";
  return `${compliance.passPercent}%`;
}

/** Classify a policy compliance signal into the same tiers used for scores. */
export function policyComplianceTier(
  compliance: PolicyCompliance
): "excellent" | "good" | "warning" | "critical" | "unknown" {
  if (!compliance) return "unknown";
  return scoreTier(compliance.passPercent);
}

/** Format a monthly cost signal for display, e.g. "$18.42/mo" or "Unknown". */
export function costLabel(cost: CostSummary): string {
  if (!cost) return "Unknown";
  return `$${cost.amount.toFixed(2)}/mo`;
}

/** Format a deployment health signal for display, e.g. "Synced · Healthy" or "Unknown". */
export function deploymentHealthLabel(health: DeploymentHealth): string {
  if (!health) return "Unknown";
  return `${health.syncStatus} · ${health.health}`;
}

/** Classify a deployment health signal into the same tiers used for scores. */
export function deploymentHealthTier(
  health: DeploymentHealth
): "excellent" | "warning" | "critical" | "unknown" {
  if (!health) return "unknown";
  return health.health.toLowerCase() === "healthy" ? "excellent" : "warning";
}

/** Compute aggregate stats, including platform-as-product adoption (pave-cli vs manual). */
export function computeStats(services: CatalogService[]) {
  const total = services.length;
  const avgScore =
    total > 0
      ? Math.round(services.reduce((sum, s) => sum + s.scorecard.overallScore, 0) / total)
      : 0;
  const passing = services.filter((s) => s.scorecard.overallScore >= 70).length;
  const totalCriteria = services.reduce((sum, s) => sum + s.scorecard.criteria.length, 0);
  const passingCriteria = services.reduce(
    (sum, s) => sum + s.scorecard.criteria.filter((c) => c.status === "passing").length,
    0
  );
  const createdViaPave = services.filter((s) => s.createdVia === "pave-cli").length;
  const adoptionPct = total > 0 ? Math.round((createdViaPave / total) * 100) : 0;
  return { total, avgScore, passing, totalCriteria, passingCriteria, createdViaPave, adoptionPct };
}

/* ────────────────────────────── Demo data ──────────────────────────── */
// Synthetic rows for illustrating table performance/virtualization at scale.
// Always disclosed in the UI via the `isDemo` flag — never merged silently
// with real catalog entries.

const DEMO_TEAMS = [
  "team-payments",
  "team-checkout",
  "team-growth",
  "team-platform",
  "team-data",
  "team-identity",
];
const DEMO_ADJECTIVES = [
  "orbit",
  "signal",
  "harbor",
  "cinder",
  "quartz",
  "lumen",
  "delta",
  "vertex",
  "coral",
  "atlas",
];
const DEMO_NOUNS = [
  "gateway",
  "ledger",
  "router",
  "worker",
  "index",
  "queue",
  "cache",
  "notifier",
  "scheduler",
  "billing",
];
const DEMO_LIFECYCLES = ["production", "staging", "experimental", "deprecated"];
const DEMO_TIERS: Tier[] = ["tier-1", "tier-2", "tier-3"];
const DEMO_RUNTIMES = ["go", "node", "python"];

/** Deterministic pseudo-random generator so demo data is stable across renders/tests. */
function mulberry32(seed: number) {
  return function () {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function generateDemoServices(count: number, seed = 42): CatalogService[] {
  const rand = mulberry32(seed);
  const pick = <T>(arr: T[]): T => arr[Math.floor(rand() * arr.length)];

  return Array.from({ length: count }, (_, i) => {
    const name = `${pick(DEMO_ADJECTIVES)}-${pick(DEMO_NOUNS)}-${i}`;
    const overallScore = Math.floor(rand() * 101);
    const createdVia: CreatedVia = rand() > 0.35 ? "pave-cli" : "manual";
    return {
      id: `demo-${name}`,
      name,
      description: "Synthetic demo row generated for table-scale illustration.",
      owner: pick(DEMO_TEAMS),
      team: pick(DEMO_TEAMS),
      system: "pavestack",
      repoUrl: "https://github.com/pavestack/pavestack",
      repoPath: `services/${name}`,
      lifecycle: pick(DEMO_LIFECYCLES),
      tier: pick(DEMO_TIERS),
      runtime: pick(DEMO_RUNTIMES),
      exposure: rand() > 0.5 ? "internal" : "public",
      database: rand() > 0.5,
      createdVia,
      environments: {
        dev: { status: "synced", health: "healthy", imageTag: `0.${i % 20}.0` },
        prod: {
          status: rand() > 0.15 ? "synced" : "outOfSync",
          health: rand() > 0.1 ? "healthy" : "degraded",
          imageTag: `0.${i % 20}.0`,
        },
      },
      scorecard: {
        overallScore,
        criteria: [
          {
            key: "security_scan_passing",
            label: "Security Scan Passing",
            status: rand() > 0.2 ? "passing" : "failing",
            weight: 30,
            evidence: null,
          },
          {
            key: "docs_present",
            label: "Docs Present",
            status: rand() > 0.3 ? "passing" : "failing",
            weight: 20,
            evidence: null,
          },
          {
            key: "health_endpoint_configured",
            label: "Health Endpoint Configured",
            status: rand() > 0.15 ? "passing" : "failing",
            weight: 25,
            evidence: null,
          },
          {
            key: "container_non_root",
            label: "Container Non Root",
            status: rand() > 0.1 ? "passing" : "failing",
            weight: 15,
            evidence: null,
          },
          {
            key: "gitops_manifests",
            label: "Gitops Manifests",
            status: rand() > 0.1 ? "passing" : "failing",
            weight: 10,
            evidence: null,
          },
        ],
      },
      isDemo: true,
    };
  });
}

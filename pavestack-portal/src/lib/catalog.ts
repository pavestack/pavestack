export type CatalogCriteria = {
  key: string;
  status: string;
  weight: number;
};

export type EnvironmentState = {
  status: string;
  health: string;
};

export type CatalogService = {
  id: string;
  name: string;
  description: string;
  owner: string;
  repo: string;
  repoPath: string;
  lifecycle: string;
  environments: Record<string, EnvironmentState>;
  scorecard: {
    overall: number;
    criteria: CatalogCriteria[];
  };
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

/** Get HSL color for a score */
export function scoreColor(score: number): string {
  if (score >= 90) return "#3fb950";
  if (score >= 70) return "#58a6ff";
  if (score >= 50) return "#d29922";
  return "#f85149";
}

/** Calculate the SVG stroke-dashoffset for a circular progress ring */
export function scoreDashOffset(score: number, circumference: number): number {
  return circumference - (score / 100) * circumference;
}

/** Filter services by search query */
export function filterServices(services: CatalogService[], query: string): CatalogService[] {
  if (!query.trim()) return services;
  const lower = query.toLowerCase();
  return services.filter(
    (s) =>
      s.name.toLowerCase().includes(lower) ||
      s.owner.toLowerCase().includes(lower) ||
      s.description.toLowerCase().includes(lower)
  );
}

/** Sort services by various criteria */
export type SortKey = "name" | "score" | "owner";

export function sortServices(services: CatalogService[], key: SortKey): CatalogService[] {
  return [...services].sort((a, b) => {
    switch (key) {
      case "score":
        return b.scorecard.overall - a.scorecard.overall;
      case "owner":
        return a.owner.localeCompare(b.owner);
      default:
        return a.name.localeCompare(b.name);
    }
  });
}

/** Compute aggregate stats */
export function computeStats(services: CatalogService[]) {
  const total = services.length;
  const avgScore =
    total > 0
      ? Math.round(services.reduce((sum, s) => sum + s.scorecard.overall, 0) / total)
      : 0;
  const passing = services.filter((s) => s.scorecard.overall >= 70).length;
  const totalCriteria = services.reduce((sum, s) => sum + s.scorecard.criteria.length, 0);
  const passingCriteria = services.reduce(
    (sum, s) => sum + s.scorecard.criteria.filter((c) => c.status === "passing").length,
    0
  );
  return { total, avgScore, passing, totalCriteria, passingCriteria };
}

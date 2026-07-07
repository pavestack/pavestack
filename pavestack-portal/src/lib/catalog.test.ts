import { describe, test, expect, vi, afterEach } from "vitest";
import {
  scoreTier,
  scoreColor,
  scoreDashOffset,
  filterServices,
  sortServices,
  computeStats,
  loadCatalog,
  policyComplianceLabel,
  policyComplianceTier,
  costLabel,
  deploymentHealthLabel,
  deploymentHealthTier,
  CatalogService
} from "./catalog";

const mockServices: CatalogService[] = [
  {
    id: "service-a",
    name: "Service A",
    description: "First service of project",
    owner: "team-alpha",
    repo: "https://github.com/org/service-a",
    repoPath: "services/service-a",
    lifecycle: "production",
    environments: {
      dev: { status: "synced", health: "healthy" },
      prod: { status: "synced", health: "healthy" }
    },
    scorecard: {
      overall: 95,
      criteria: [
        { key: "security_scan_passing", status: "passing", weight: 50 },
        { key: "docs_present", status: "passing", weight: 50 }
      ]
    }
  },
  {
    id: "service-b",
    name: "Service B",
    description: "Second project service",
    owner: "team-beta",
    repo: "https://github.com/org/service-b",
    repoPath: "services/service-b",
    lifecycle: "staging",
    environments: {
      dev: { status: "synced", health: "healthy" }
    },
    scorecard: {
      overall: 75,
      criteria: [
        { key: "security_scan_passing", status: "passing", weight: 50 },
        { key: "docs_present", status: "failing", weight: 50 }
      ]
    }
  },
  {
    id: "service-c",
    name: "Service C",
    description: "Third deprecated service",
    owner: "team-alpha",
    repo: "https://github.com/org/service-c",
    repoPath: "services/service-c",
    lifecycle: "deprecated",
    environments: {},
    scorecard: {
      overall: 45,
      criteria: [
        { key: "security_scan_passing", status: "failing", weight: 50 },
        { key: "docs_present", status: "failing", weight: 50 }
      ]
    }
  }
];

describe("Catalog Helpers", () => {
  describe("scoreTier", () => {
    test("classifies score into tiers", () => {
      expect(scoreTier(95)).toBe("excellent");
      expect(scoreTier(90)).toBe("excellent");
      expect(scoreTier(89)).toBe("good");
      expect(scoreTier(70)).toBe("good");
      expect(scoreTier(69)).toBe("warning");
      expect(scoreTier(50)).toBe("warning");
      expect(scoreTier(49)).toBe("critical");
      expect(scoreTier(0)).toBe("critical");
    });
  });

  describe("scoreColor", () => {
    test("returns correct colors", () => {
      expect(scoreColor(95)).toBe("#3fb950");
      expect(scoreColor(75)).toBe("#58a6ff");
      expect(scoreColor(55)).toBe("#d29922");
      expect(scoreColor(35)).toBe("#f85149");
    });
  });

  describe("scoreDashOffset", () => {
    test("calculates dashoffset accurately", () => {
      const circ = 100;
      expect(scoreDashOffset(100, circ)).toBe(0);
      expect(scoreDashOffset(50, circ)).toBe(50);
      expect(scoreDashOffset(0, circ)).toBe(100);
      expect(scoreDashOffset(25, circ)).toBe(75);
    });
  });

  describe("filterServices", () => {
    test("returns all services if query is empty", () => {
      expect(filterServices(mockServices, "")).toEqual(mockServices);
      expect(filterServices(mockServices, "   ")).toEqual(mockServices);
    });

    test("filters services by name (case-insensitive)", () => {
      const res = filterServices(mockServices, "service a");
      expect(res).toHaveLength(1);
      expect(res[0].id).toBe("service-a");
    });

    test("filters services by owner (case-insensitive)", () => {
      const res = filterServices(mockServices, "team-alpha");
      expect(res).toHaveLength(2);
      expect(res.map(s => s.id)).toContain("service-a");
      expect(res.map(s => s.id)).toContain("service-c");
    });

    test("filters services by description (case-insensitive)", () => {
      const res = filterServices(mockServices, "second");
      expect(res).toHaveLength(1);
      expect(res[0].id).toBe("service-b");
    });
  });

  describe("sortServices", () => {
    test("sorts services by name", () => {
      const sorted = sortServices(mockServices, "name");
      expect(sorted[0].id).toBe("service-a");
      expect(sorted[1].id).toBe("service-b");
      expect(sorted[2].id).toBe("service-c");
    });

    test("sorts services by score (descending)", () => {
      // Input is mockServices: [95, 75, 45]
      const sorted = sortServices([mockServices[2], mockServices[0], mockServices[1]], "score");
      expect(sorted[0].id).toBe("service-a"); // 95
      expect(sorted[1].id).toBe("service-b"); // 75
      expect(sorted[2].id).toBe("service-c"); // 45
    });

    test("sorts services by owner", () => {
      const sorted = sortServices(mockServices, "owner");
      // team-alpha, team-alpha, team-beta -> Service A (alpha), Service C (alpha), Service B (beta)
      expect(sorted[0].id).toBe("service-a");
      expect(sorted[1].id).toBe("service-c");
      expect(sorted[2].id).toBe("service-b");
    });
  });

  describe("computeStats", () => {
    test("calculates statistics correctly", () => {
      const stats = computeStats(mockServices);
      expect(stats.total).toBe(3);
      expect(stats.avgScore).toBe(Math.round((95 + 75 + 45) / 3)); // 72
      expect(stats.passing).toBe(2); // A (95) and B (75) are >= 70
      expect(stats.totalCriteria).toBe(6); // 2 + 2 + 2
      expect(stats.passingCriteria).toBe(3); // 2 passing in A, 1 passing in B, 0 in C
    });

    test("returns zero stats for empty list", () => {
      const stats = computeStats([]);
      expect(stats.total).toBe(0);
      expect(stats.avgScore).toBe(0);
      expect(stats.passing).toBe(0);
      expect(stats.totalCriteria).toBe(0);
      expect(stats.passingCriteria).toBe(0);
    });
  });

  describe("policyComplianceLabel / policyComplianceTier", () => {
    test("formats a known compliance signal and tiers it like a score", () => {
      const compliance = { pass: 24, warn: 2, fail: 0, passPercent: 92 };
      expect(policyComplianceLabel(compliance)).toBe("92%");
      expect(policyComplianceTier(compliance)).toBe("excellent");
    });

    test("falls back to unknown when the artifact/service entry is absent", () => {
      expect(policyComplianceLabel(null)).toBe("Unknown");
      expect(policyComplianceTier(null)).toBe("unknown");
    });
  });

  describe("costLabel", () => {
    test("formats a known monthly cost", () => {
      expect(costLabel({ amount: 18.42, currency: "USD" })).toBe("$18.42/mo");
    });

    test("falls back to unknown when the artifact/service entry is absent", () => {
      expect(costLabel(null)).toBe("Unknown");
    });
  });

  describe("deploymentHealthLabel / deploymentHealthTier", () => {
    test("formats and tiers a healthy deployment", () => {
      const health = { syncStatus: "Synced", health: "Healthy", lastSyncAt: "2026-07-05T09:08:11Z" };
      expect(deploymentHealthLabel(health)).toBe("Synced · Healthy");
      expect(deploymentHealthTier(health)).toBe("excellent");
    });

    test("tiers a non-healthy deployment as a warning", () => {
      const health = { syncStatus: "OutOfSync", health: "Degraded", lastSyncAt: null };
      expect(deploymentHealthTier(health)).toBe("warning");
    });

    test("falls back to unknown when the artifact/service entry is absent", () => {
      expect(deploymentHealthLabel(null)).toBe("Unknown");
      expect(deploymentHealthTier(null)).toBe("unknown");
    });
  });

  describe("CatalogService with report-derived signals", () => {
    test("services with populated signals still filter, sort and aggregate correctly", () => {
      const enriched: CatalogService = {
        ...mockServices[0],
        policyCompliance: { dev: { pass: 24, warn: 2, fail: 0, passPercent: 92 }, prod: null },
        costPerMonth: { dev: { amount: 18.42, currency: "USD" }, prod: null },
        deploymentHealth: {
          dev: { syncStatus: "Synced", health: "Healthy", lastSyncAt: "2026-07-05T09:08:11Z" },
          prod: null
        },
        api: {
          title: "Service A",
          version: "1.0.0",
          endpoints: [{ method: "GET", path: "/health", summary: "Liveness probe" }]
        }
      };

      const services = [enriched, mockServices[1]];
      expect(filterServices(services, "service a")).toHaveLength(1);
      expect(sortServices(services, "score")[0].id).toBe("service-a");
      expect(computeStats(services).total).toBe(2);
    });

    test("services without report artifacts (older catalogs) remain valid and unaffected", () => {
      // mockServices entries have no policyCompliance/costPerMonth/deploymentHealth/api
      // fields at all — these must stay optional so older generated catalogs keep working.
      expect(computeStats(mockServices).total).toBe(3);
    });
  });

  describe("loadCatalog", () => {
    afterEach(() => {
      vi.restoreAllMocks();
    });

    test("successfully loads catalog via fetch", async () => {
      const mockCatalogData = {
        generatedAt: "2026-06-19T00:00:00Z",
        services: [mockServices[0]]
      };

      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockCatalogData
      });
      vi.stubGlobal("fetch", fetchMock);

      const catalog = await loadCatalog();
      expect(fetchMock).toHaveBeenCalledWith("/catalog.json");
      expect(catalog).toEqual(mockCatalogData);
    });

    test("throws an error when response is not ok", async () => {
      const fetchMock = vi.fn().mockResolvedValue({
        ok: false
      });
      vi.stubGlobal("fetch", fetchMock);

      await expect(loadCatalog()).rejects.toThrow("Failed to load catalog");
    });
  });
});

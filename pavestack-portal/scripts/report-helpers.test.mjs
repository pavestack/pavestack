import { describe, test, expect } from "vitest";
import {
  computePolicyCompliance,
  computeCostPerMonth,
  computeDeploymentHealth,
  parseOpenApiDoc,
} from "./report-helpers.mjs";

describe("computePolicyCompliance", () => {
  test("computes pass percent from a matching namespace entry", () => {
    const report = { namespaces: { "service-a": { pass: 18, warn: 2, fail: 0 } } };
    expect(computePolicyCompliance(report, "service-a")).toEqual({
      pass: 18,
      warn: 2,
      fail: 0,
      passPercent: 90,
    });
  });

  test("returns null when the report is missing (unknown)", () => {
    expect(computePolicyCompliance(null, "service-a")).toBeNull();
    expect(computePolicyCompliance(undefined, "service-a")).toBeNull();
  });

  test("returns null when the service has no entry in the report", () => {
    const report = { namespaces: { "other-service": { pass: 1, warn: 0, fail: 0 } } };
    expect(computePolicyCompliance(report, "service-a")).toBeNull();
  });

  test("handles a namespace with zero checks without dividing by zero", () => {
    const report = { namespaces: { "service-a": { pass: 0, warn: 0, fail: 0 } } };
    expect(computePolicyCompliance(report, "service-a")).toEqual({
      pass: 0,
      warn: 0,
      fail: 0,
      passPercent: 0,
    });
  });
});

describe("computeCostPerMonth", () => {
  test("reads monthly cost and currency for a matching namespace", () => {
    const report = { currency: "USD", namespaces: { "service-a": { monthlyCost: 18.42 } } };
    expect(computeCostPerMonth(report, "service-a")).toEqual({ amount: 18.42, currency: "USD" });
  });

  test("defaults currency to USD when not specified", () => {
    const report = { namespaces: { "service-a": { monthlyCost: 5 } } };
    expect(computeCostPerMonth(report, "service-a")).toEqual({ amount: 5, currency: "USD" });
  });

  test("returns null when the report is missing (unknown)", () => {
    expect(computeCostPerMonth(null, "service-a")).toBeNull();
  });

  test("returns null when the service has no cost entry", () => {
    const report = { namespaces: {} };
    expect(computeCostPerMonth(report, "service-a")).toBeNull();
  });
});

describe("computeDeploymentHealth", () => {
  test("reads sync status, health and last sync time for a matching app", () => {
    const report = {
      apps: {
        "service-a": {
          syncStatus: "Synced",
          health: "Healthy",
          lastSyncAt: "2026-07-05T09:08:11Z",
        },
      },
    };
    expect(computeDeploymentHealth(report, "service-a")).toEqual({
      syncStatus: "Synced",
      health: "Healthy",
      lastSyncAt: "2026-07-05T09:08:11Z",
    });
  });

  test("returns null when the report is missing (unknown)", () => {
    expect(computeDeploymentHealth(undefined, "service-a")).toBeNull();
  });

  test("returns null when the app has no entry in the report", () => {
    const report = { apps: {} };
    expect(computeDeploymentHealth(report, "service-a")).toBeNull();
  });
});

describe("parseOpenApiDoc", () => {
  test("extracts title, version and endpoints from a spec", () => {
    const doc = {
      info: { title: "Service Template API", version: "1.0.0" },
      paths: {
        "/healthz": { get: { summary: "Health check" } },
        "/widgets": {
          get: { summary: "List widgets" },
          post: { operationId: "createWidget" },
        },
      },
    };
    expect(parseOpenApiDoc(doc)).toEqual({
      title: "Service Template API",
      version: "1.0.0",
      endpoints: [
        { method: "GET", path: "/healthz", summary: "Health check" },
        { method: "GET", path: "/widgets", summary: "List widgets" },
        { method: "POST", path: "/widgets", summary: "createWidget" },
      ],
    });
  });

  test("returns null for a missing document", () => {
    expect(parseOpenApiDoc(null)).toBeNull();
    expect(parseOpenApiDoc(undefined)).toBeNull();
  });

  test("tolerates a document with no paths", () => {
    expect(parseOpenApiDoc({ info: { title: "Empty", version: "0.1.0" } })).toEqual({
      title: "Empty",
      version: "0.1.0",
      endpoints: [],
    });
  });

  test("ignores non-HTTP-method keys under a path (e.g. parameters)", () => {
    const doc = {
      info: {},
      paths: {
        "/items/{id}": {
          parameters: [{ name: "id", in: "path" }],
          get: { summary: "Get item" },
        },
      },
    };
    expect(parseOpenApiDoc(doc).endpoints).toEqual([
      { method: "GET", path: "/items/{id}", summary: "Get item" },
    ]);
  });
});

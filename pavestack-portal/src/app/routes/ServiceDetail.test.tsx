import React from "react";
import { describe, test, expect, vi, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { CatalogProvider } from "../CatalogContext";
import { ServiceDetail } from "./ServiceDetail";
import type { Catalog, CatalogService } from "../../lib/catalog";

const baseService: CatalogService = {
  id: "payments-api",
  name: "payments-api",
  description: "Handles payment processing",
  owner: "team-payments",
  team: "team-payments",
  system: "pavestack",
  repoUrl: "https://github.com/org/payments-api",
  repoPath: "services/payments-api",
  lifecycle: "production",
  tier: "tier-1",
  runtime: "go",
  exposure: "internal",
  database: true,
  createdVia: "pave-cli",
  environments: {
    dev: { status: "synced", health: "healthy", imageTag: "1.0.0" },
  },
  scorecard: { overallScore: 90, criteria: [] },
};

function catalogWith(service: CatalogService): Catalog {
  return { generatedAt: new Date(0).toISOString(), services: [service] };
}

function mockFetch(catalog: Catalog, costResponse?: unknown, costOk = true) {
  return vi.fn((url: string | URL) => {
    const href = url.toString();
    if (href.includes("/catalog.json")) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(catalog) });
    }
    if (href.includes("/cost-estimate")) {
      return Promise.resolve({
        ok: costOk,
        status: costOk ? 200 : 502,
        json: () => Promise.resolve(costOk ? costResponse : { error: "unreachable" }),
      });
    }
    return Promise.reject(new Error(`unexpected fetch: ${href}`));
  });
}

function renderAt(name: string) {
  return render(
    <MemoryRouter initialEntries={[`/services/${name}`]}>
      <CatalogProvider>
        <Routes>
          <Route path="/services/:name" element={<ServiceDetail />} />
        </Routes>
      </CatalogProvider>
    </MemoryRouter>
  );
}

describe("ServiceDetail", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  test("renders service metadata and fetches a live cost estimate", async () => {
    const costResponse = {
      monthlyUsdLow: 50,
      monthlyUsdHigh: 100,
      currency: "USD",
      breakdown: [{ item: "compute", monthlyUsd: 75 }],
      disclaimer: "Estimate only.",
    };
    vi.stubGlobal("fetch", mockFetch(catalogWith(baseService), costResponse));

    renderAt("payments-api");

    expect(await screen.findByText("payments-api")).toBeInTheDocument();
    expect(screen.getByText("team-payments")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/\$50–\$100/)).toBeInTheDocument();
    });
    expect(screen.getByText("compute")).toBeInTheDocument();
  });

  test("shows an empty state for an unknown service name", async () => {
    vi.stubGlobal("fetch", mockFetch(catalogWith(baseService)));

    renderAt("does-not-exist");

    expect(await screen.findByText("Service not found")).toBeInTheDocument();
  });

  test("shows a message instead of a cost estimate when tier/exposure are unset", async () => {
    const noTierService: CatalogService = { ...baseService, tier: null, exposure: null };
    vi.stubGlobal("fetch", mockFetch(catalogWith(noTierService)));

    renderAt("payments-api");

    expect(await screen.findByText("payments-api")).toBeInTheDocument();
    expect(screen.getByText(/Tier and exposure aren't set/)).toBeInTheDocument();
  });

  test("shows an inline error when the cost estimate call fails", async () => {
    vi.stubGlobal("fetch", mockFetch(catalogWith(baseService), undefined, false));

    renderAt("payments-api");

    expect(await screen.findByText("payments-api")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText(/Backend unreachable/)).toBeInTheDocument();
    });
  });
});

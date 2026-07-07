import React from "react";
import { describe, test, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CatalogProvider } from "../CatalogContext";
import { Scorecards } from "./Scorecards";
import type { Catalog, CatalogService } from "../../lib/catalog";

function service(name: string, score: number): CatalogService {
  return {
    id: name,
    name,
    description: "",
    owner: "team-a",
    team: "team-a",
    system: "pavestack",
    repoUrl: "https://github.com/org/" + name,
    repoPath: "services/" + name,
    lifecycle: "production",
    tier: "tier-2",
    runtime: "go",
    exposure: "internal",
    database: false,
    createdVia: "pave-cli",
    environments: {},
    scorecard: { overallScore: score, criteria: [] },
  };
}

function mockFetch(catalog: Catalog) {
  return vi.fn(() => Promise.resolve({ ok: true, json: () => Promise.resolve(catalog) }));
}

function renderScorecards() {
  return render(
    <MemoryRouter>
      <CatalogProvider>
        <Scorecards />
      </CatalogProvider>
    </MemoryRouter>
  );
}

describe("Scorecards", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  test("ranks services by score, highest first by default", async () => {
    const catalog: Catalog = {
      generatedAt: new Date(0).toISOString(),
      services: [service("low-score", 40), service("high-score", 95), service("mid-score", 70)],
    };
    vi.stubGlobal("fetch", mockFetch(catalog));

    renderScorecards();

    await screen.findByText("high-score");
    const rows = screen.getAllByRole("row").slice(1); // drop header row
    expect(rows[0]).toHaveTextContent("high-score");
    expect(rows[1]).toHaveTextContent("mid-score");
    expect(rows[2]).toHaveTextContent("low-score");
  });

  test("toggling sort reverses the ranking", async () => {
    const catalog: Catalog = {
      generatedAt: new Date(0).toISOString(),
      services: [service("low-score", 40), service("high-score", 95)],
    };
    vi.stubGlobal("fetch", mockFetch(catalog));

    renderScorecards();
    await screen.findByText("high-score");

    fireEvent.click(screen.getByRole("button", { name: /Score/ }));

    await waitFor(() => {
      const rows = screen.getAllByRole("row").slice(1);
      expect(rows[0]).toHaveTextContent("low-score");
    });
  });

  test("shows a perfect-score banner when a service scores 100", async () => {
    const catalog: Catalog = {
      generatedAt: new Date(0).toISOString(),
      services: [service("perfect-service", 100)],
    };
    vi.stubGlobal("fetch", mockFetch(catalog));

    renderScorecards();

    expect(await screen.findByText(/perfect 100/)).toBeInTheDocument();
  });

  test("shows an empty state when the catalog has no services", async () => {
    const catalog: Catalog = { generatedAt: new Date(0).toISOString(), services: [] };
    vi.stubGlobal("fetch", mockFetch(catalog));

    renderScorecards();

    expect(await screen.findByText("No services scored yet")).toBeInTheDocument();
  });
});

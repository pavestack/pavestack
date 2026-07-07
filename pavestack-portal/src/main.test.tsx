import React from "react";
import { describe, test, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import {
  ScoreRing,
  StatCard,
  CriteriaRow,
  EnvBadge,
  SortSelect,
  ServiceCard,
} from "./app/components";
import { MemoryRouter } from "react-router-dom";
import { App } from "./app";
import { CatalogService } from "./lib/catalog";

const mockService: CatalogService = {
  id: "test-service",
  name: "Test Service",
  description: "A service for testing UI components",
  owner: "team-testing",
  team: "team-testing",
  system: "pavestack",
  repoUrl: "https://github.com/org/test-service",
  repoPath: "services/test-service",
  lifecycle: "production",
  tier: "tier-2",
  runtime: "go",
  exposure: "internal",
  database: false,
  createdVia: "pave-cli",
  environments: {
    dev: { status: "synced", health: "healthy", imageTag: "1.2.0" },
    prod: { status: "outOfSync", health: "unhealthy", imageTag: "1.1.0" },
  },
  scorecard: {
    overallScore: 85,
    criteria: [
      {
        key: "security_scan_passing",
        label: "Security Scan Passing",
        status: "passing",
        weight: 60,
      },
      { key: "docs_present", label: "Docs Present", status: "failing", weight: 40 },
    ],
  },
};

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("UI Components", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("ScoreRing", () => {
    test("renders the correct score value", () => {
      render(<ScoreRing score={85} />);
      expect(screen.getByText("85")).toBeInTheDocument();
    });

    test("applies the theme-aware color token for the score tier", () => {
      const { container } = render(<ScoreRing score={95} size={100} />);
      const fillCircle = container.querySelector(".score-ring-fill");
      expect(fillCircle).toBeInTheDocument();
      expect(fillCircle).toHaveStyle({ stroke: "var(--success)" });
    });
  });

  describe("StatCard", () => {
    test("renders the label, value, and subtext", () => {
      render(
        <StatCard
          icon={<span>icon-placeholder</span>}
          label="Test Label"
          value="Test Value"
          subtext="Test Subtext"
        />
      );
      expect(screen.getByText("Test Label")).toBeInTheDocument();
      expect(screen.getByText("Test Value")).toBeInTheDocument();
      expect(screen.getByText("Test Subtext")).toBeInTheDocument();
      expect(screen.getByText("icon-placeholder")).toBeInTheDocument();
    });
  });

  describe("CriteriaRow", () => {
    test("renders passing criteria correctly", () => {
      render(
        <CriteriaRow
          item={{ key: "docs_present", label: "Docs Present", status: "passing", weight: 30 }}
        />
      );
      expect(screen.getByText("Docs Present")).toBeInTheDocument();
      expect(screen.getByText("passing")).toBeInTheDocument();
      expect(screen.getByText("30%")).toBeInTheDocument();
    });

    test("renders failing criteria correctly", () => {
      render(
        <CriteriaRow
          item={{
            key: "security_scan_passing",
            label: "Security Scan Passing",
            status: "failing",
            weight: 70,
          }}
        />
      );
      expect(screen.getByText("Security Scan Passing")).toBeInTheDocument();
      expect(screen.getByText("failing")).toBeInTheDocument();
      expect(screen.getByText("70%")).toBeInTheDocument();
    });
  });

  describe("EnvBadge", () => {
    test("renders synced/healthy environment as success badge style", () => {
      const { container } = render(<EnvBadge env="dev" status="synced" health="healthy" />);
      expect(screen.getByText("dev")).toBeInTheDocument();
      expect(screen.getByText("synced · healthy")).toBeInTheDocument();
      expect(container.firstChild).toHaveClass("bg-pave-success/5");
    });

    test("renders unhealthy environment with warning badge style", () => {
      const { container } = render(<EnvBadge env="prod" status="outOfSync" health="unhealthy" />);
      expect(screen.getByText("prod")).toBeInTheDocument();
      expect(screen.getByText("outOfSync · unhealthy")).toBeInTheDocument();
      expect(container.firstChild).toHaveClass("bg-pave-warning/5");
    });

    test("shows image tag when provided", () => {
      render(<EnvBadge env="dev" status="synced" health="healthy" imageTag="1.2.0" />);
      expect(screen.getByText("1.2.0")).toBeInTheDocument();
    });
  });

  describe("SortSelect", () => {
    test("renders options and triggers onChange callback", () => {
      const handleChange = vi.fn();
      render(<SortSelect value="name" onChange={handleChange} />);

      const select = screen.getByRole("combobox") as HTMLSelectElement;
      expect(select.value).toBe("name");

      fireEvent.change(select, { target: { value: "score" } });
      expect(handleChange).toHaveBeenCalledWith("score");
    });
  });

  describe("ServiceCard", () => {
    test("renders basic service details", () => {
      renderWithRouter(<ServiceCard service={mockService} />);
      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("good")).toBeInTheDocument(); // score 85 is "good"
      expect(screen.getByText("A service for testing UI components")).toBeInTheDocument();
      expect(screen.getByText("team-testing")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("services/test-service")).toBeInTheDocument();
      expect(screen.getByText("pave-cli")).toBeInTheDocument();
    });

    test("toggles scorecard details section when clicked", () => {
      renderWithRouter(<ServiceCard service={mockService} />);
      expect(screen.queryByText("Security Scan Passing")).not.toBeInTheDocument();

      const button = screen.getByRole("button", { name: /Scorecard details/ });
      expect(button).toBeInTheDocument();
      expect(button).toHaveAttribute("aria-expanded", "false");

      fireEvent.click(button);
      expect(button).toHaveAttribute("aria-expanded", "true");
      expect(screen.getByText("Security Scan Passing")).toBeInTheDocument();
      expect(screen.getByText("Docs Present")).toBeInTheDocument();

      fireEvent.click(button);
      expect(button).toHaveAttribute("aria-expanded", "false");
      expect(screen.queryByText("Security Scan Passing")).not.toBeInTheDocument();
    });
  });

  describe("App", () => {
    test("shows a skeleton loading state initially, then renders catalog content", async () => {
      const mockCatalogData = {
        generatedAt: "2026-06-19T00:00:00Z",
        services: [mockService],
      };

      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockCatalogData,
      });
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByRole("status", { name: /loading/i })).not.toBeInTheDocument();
      });

      expect(screen.getByText("Service catalog")).toBeInTheDocument();
      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("Registered in catalog")).toBeInTheDocument();
    });

    test("shows an inline error and empty state when the catalog fetch fails", async () => {
      const fetchMock = vi.fn().mockRejectedValue(new Error("Network Error"));
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      await waitFor(() => {
        expect(screen.queryByRole("status", { name: /loading/i })).not.toBeInTheDocument();
      });

      expect(screen.getByRole("alert")).toHaveTextContent(/Failed to load catalog/);
      expect(screen.getByText("No services registered yet")).toBeInTheDocument();
    });

    test("filters services dynamically based on search input", async () => {
      const mockCatalogData = {
        generatedAt: "2026-06-19T00:00:00Z",
        services: [
          mockService,
          {
            ...mockService,
            id: "other-service",
            name: "Other Unique Name",
            description: "Totally different desc",
          },
        ],
      };

      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockCatalogData,
      });
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      await waitFor(() => {
        expect(screen.queryByRole("status", { name: /loading/i })).not.toBeInTheDocument();
      });

      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();

      const searchInput = screen.getByPlaceholderText(
        "Search services by name, owner, or description…"
      );
      fireEvent.change(searchInput, { target: { value: "Unique" } });

      expect(screen.queryByText("Test Service")).not.toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();

      fireEvent.change(searchInput, { target: { value: "" } });
      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();
    });
  });
});

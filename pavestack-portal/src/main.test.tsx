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
  App
} from "./main";
import { CatalogService } from "./lib/catalog";

const mockService: CatalogService = {
  id: "test-service",
  name: "Test Service",
  description: "A service for testing UI components",
  owner: "team-testing",
  repo: "https://github.com/org/test-service",
  repoPath: "services/test-service",
  lifecycle: "production",
  environments: {
    dev: { status: "synced", health: "healthy" },
    prod: { status: "outOfSync", health: "unhealthy" }
  },
  scorecard: {
    overall: 85,
    criteria: [
      { key: "security_scan_passing", status: "passing", weight: 60 },
      { key: "docs_present", status: "failing", weight: 40 }
    ]
  }
};

describe("UI Components in main.tsx", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("ScoreRing", () => {
    test("renders the correct score value", () => {
      render(<ScoreRing score={85} />);
      expect(screen.getByText("85")).toBeInTheDocument();
    });

    test("applies styling corresponding to the score color", () => {
      const { container } = render(<ScoreRing score={95} size={100} />);
      const fillCircle = container.querySelector(".score-ring-fill");
      expect(fillCircle).toBeInTheDocument();
      // Score 95 should be excellent, color #3fb950
      expect(fillCircle).toHaveStyle({ stroke: "#3fb950" });
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
          delay={0}
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
      render(<CriteriaRow item={{ key: "docs_present", status: "passing", weight: 30 }} />);
      expect(screen.getByText("docs present")).toBeInTheDocument();
      expect(screen.getByText("passing")).toBeInTheDocument();
      expect(screen.getByText("30%")).toBeInTheDocument();
    });

    test("renders failing criteria correctly", () => {
      render(<CriteriaRow item={{ key: "security_scan_passing", status: "failing", weight: 70 }} />);
      expect(screen.getByText("security scan passing")).toBeInTheDocument();
      expect(screen.getByText("failing")).toBeInTheDocument();
      expect(screen.getByText("70%")).toBeInTheDocument();
    });
  });

  describe("EnvBadge", () => {
    test("renders synced/healthy environment as success badge style", () => {
      const { container } = render(<EnvBadge env="dev" status="synced" health="healthy" />);
      expect(screen.getByText("dev")).toBeInTheDocument();
      expect(screen.getByText("synced · healthy")).toBeInTheDocument();
      // allGood true -> contains class bg-pave-success/5
      expect(container.firstChild).toHaveClass("bg-pave-success/5");
    });

    test("renders unhealthy environment with warning badge style", () => {
      const { container } = render(<EnvBadge env="prod" status="outOfSync" health="unhealthy" />);
      expect(screen.getByText("prod")).toBeInTheDocument();
      expect(screen.getByText("outOfSync · unhealthy")).toBeInTheDocument();
      // allGood false -> contains class bg-pave-warning/5
      expect(container.firstChild).toHaveClass("bg-pave-warning/5");
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
      render(<ServiceCard service={mockService} index={0} />);
      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("good")).toBeInTheDocument(); // score 85 is "good"
      expect(screen.getByText("A service for testing UI components")).toBeInTheDocument();
      expect(screen.getByText("team-testing")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("services/test-service")).toBeInTheDocument();
    });

    test("toggles scorecard details section when clicked", () => {
      render(<ServiceCard service={mockService} index={0} />);
      // Initially, the details should not be rendered
      expect(screen.queryByText("security scan passing")).not.toBeInTheDocument();

      const button = screen.getByRole("button", { name: /Scorecard Details/ });
      expect(button).toBeInTheDocument();
      expect(button).toHaveAttribute("aria-expanded", "false");

      // Click to expand
      fireEvent.click(button);
      expect(button).toHaveAttribute("aria-expanded", "true");
      expect(screen.getByText("security scan passing")).toBeInTheDocument();
      expect(screen.getByText("docs present")).toBeInTheDocument();

      // Click to collapse again
      fireEvent.click(button);
      expect(button).toHaveAttribute("aria-expanded", "false");
      expect(screen.queryByText("security scan passing")).not.toBeInTheDocument();
    });
  });

  describe("App", () => {
    test("shows loading state initially, then renders catalog content", async () => {
      const mockCatalogData = {
        generatedAt: "2026-06-19T00:00:00Z",
        services: [mockService]
      };

      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockCatalogData
      });
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      // Verify loading state is displayed first
      expect(screen.getByText("Loading catalog…")).toBeInTheDocument();

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText("Loading catalog…")).not.toBeInTheDocument();
      });

      // Verify stats and title are rendered
      expect(screen.getByText("Software Catalog")).toBeInTheDocument();
      expect(screen.getByText("Test Service")).toBeInTheDocument();

      // Verify stats card content
      expect(screen.getByText("Registered in catalog")).toBeInTheDocument();
      // Total services should be 1
      expect(screen.getByText("1")).toBeInTheDocument();
    });

    test("shows empty state and warning banner when fetch fails", async () => {
      const fetchMock = vi.fn().mockRejectedValue(new Error("Network Error"));
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      await waitFor(() => {
        expect(screen.queryByText("Loading catalog…")).not.toBeInTheDocument();
      });

      expect(screen.getByText("⚠ Failed to load catalog data. Showing empty state.")).toBeInTheDocument();
      expect(screen.getByText("No services registered")).toBeInTheDocument();
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
            description: "Totally different desc"
          }
        ]
      };

      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockCatalogData
      });
      vi.stubGlobal("fetch", fetchMock);

      render(<App />);

      await waitFor(() => {
        expect(screen.queryByText("Loading catalog…")).not.toBeInTheDocument();
      });

      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();

      // Search for "Unique"
      const searchInput = screen.getByPlaceholderText("Search services by name, owner, or description…");
      fireEvent.change(searchInput, { target: { value: "Unique" } });

      // Test Service should be filtered out, Other Unique Name should remain
      expect(screen.queryByText("Test Service")).not.toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();

      // Clear search
      fireEvent.change(searchInput, { target: { value: "" } });
      expect(screen.getByText("Test Service")).toBeInTheDocument();
      expect(screen.getByText("Other Unique Name")).toBeInTheDocument();
    });
  });
});

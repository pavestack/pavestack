import React from "react";
import { describe, test, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CreateServiceWizard } from "./CreateServiceWizard";

function jsonResponse(body: unknown, ok = true, status = ok ? 200 : 400) {
  return Promise.resolve({ ok, status, json: () => Promise.resolve(body) });
}

function renderWizard() {
  return render(
    <MemoryRouter>
      <CreateServiceWizard />
    </MemoryRouter>
  );
}

describe("CreateServiceWizard", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  test("blocks Continue on step 0 until name and team are valid", () => {
    renderWizard();

    fireEvent.click(screen.getByRole("button", { name: "Continue" }));
    // still on step 0: the runtime fieldset from step 1 must not be present
    // (the step tracker itself always shows a "Runtime" label, so check for
    // the step-1 fieldset specifically via its accessible group role).
    expect(screen.queryByRole("group", { name: "Runtime" })).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Service name"), { target: { value: "Not Valid!" } });
    fireEvent.click(screen.getByRole("button", { name: "Continue" }));
    expect(screen.getByText(/lowercase letters, numbers, or hyphens/)).toBeInTheDocument();
  });

  test("walks through all steps, fetches a cost estimate, and submits", async () => {
    const fetchMock = vi.fn((url: string | URL, init?: RequestInit) => {
      const href = url.toString();
      if (href.includes("/cost-estimate")) {
        return jsonResponse({
          monthlyUsdLow: 30,
          monthlyUsdHigh: 60,
          currency: "USD",
          breakdown: [],
          disclaimer: "Estimate only.",
        });
      }
      if (href.includes("/services") && init?.method === "POST") {
        return jsonResponse({ jobId: "job_abc", statusUrl: "/api/v1/jobs/job_abc" }, true, 202);
      }
      if (href.includes("/jobs/job_abc")) {
        return jsonResponse({
          jobId: "job_abc",
          status: "completed",
          steps: [],
          dryRun: true,
          prUrl: "https://github.com/org/repo/pull/1",
        });
      }
      return Promise.reject(new Error(`unexpected fetch: ${href}`));
    });
    vi.stubGlobal("fetch", fetchMock);

    renderWizard();

    fireEvent.change(screen.getByLabelText("Service name"), { target: { value: "payments" } });
    fireEvent.change(screen.getByLabelText("Owning team"), { target: { value: "team-payments" } });
    fireEvent.click(screen.getByRole("button", { name: "Continue" })); // -> step 1 (runtime)

    expect(screen.getByRole("group", { name: "Runtime" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Continue" })); // -> step 2 (shape/tier)

    expect(screen.getByRole("group", { name: "Exposure" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Continue" })); // -> step 3 (review)

    expect(screen.getByText("Review")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText(/\$30–\$60/)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Create service" }));

    await waitFor(() => {
      expect(screen.getByText(/Creating payments/)).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByRole("link", { name: /View pull request/ })).toBeInTheDocument();
    });
  });
});

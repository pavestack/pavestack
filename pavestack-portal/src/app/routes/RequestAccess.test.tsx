import React from "react";
import { describe, test, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { RequestAccess } from "./RequestAccess";
import type { AccessRequest } from "../../lib/api";

function jsonResponse(body: unknown, ok = true, status = ok ? 200 : 400) {
  return Promise.resolve({ ok, status, json: () => Promise.resolve(body) });
}

const existingRequest: AccessRequest = {
  id: "ar_1",
  requester: "alice",
  team: "team-payments",
  namespace: "payments",
  level: "write",
  reason: "on-call rotation",
  status: "pending",
  createdAt: new Date(0).toISOString(),
};

describe("RequestAccess", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  test("lists existing access requests", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => jsonResponse([existingRequest]))
    );

    render(<RequestAccess />);

    expect(await screen.findByText("payments")).toBeInTheDocument();
    expect(screen.getByText(/alice · team-payments/)).toBeInTheDocument();
  });

  test("shows an empty state when there are no requests", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => jsonResponse([]))
    );

    render(<RequestAccess />);

    expect(await screen.findByText("No access requests yet")).toBeInTheDocument();
  });

  test("submits the form and shows a pending confirmation", async () => {
    const created: AccessRequest = { ...existingRequest, id: "ar_2", status: "pending" };
    const fetchMock = vi.fn((url: string | URL, init?: RequestInit) => {
      const href = url.toString();
      if (init?.method === "POST") return jsonResponse(created, true, 201);
      return jsonResponse(href.includes("ar_2") ? [created] : []);
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<RequestAccess />);
    await screen.findByText("No access requests yet");

    fireEvent.change(screen.getByLabelText("Requester"), { target: { value: "bob" } });
    fireEvent.change(screen.getByLabelText("Team"), { target: { value: "team-infra" } });
    fireEvent.change(screen.getByLabelText("Namespace"), { target: { value: "infra" } });
    fireEvent.change(screen.getByLabelText("Reason"), { target: { value: "need write access" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit request" }));

    await waitFor(() => {
      expect(screen.getByText(/pending approval/)).toBeInTheDocument();
    });
  });

  test("shows a sign-in prompt when the API returns 401", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => jsonResponse({ error: "authentication required" }, false, 401))
    );

    render(<RequestAccess />);

    expect(await screen.findByText(/Sign in required/)).toBeInTheDocument();
  });
});

/**
 * Client for the `pave-api` backend (built separately, in Go, at
 * `pave/cmd/pave-api`). It may not be running — every call surfaces a
 * distinguishable "unreachable" failure so the UI can show an honest
 * offline state instead of faking success.
 */

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8787/api/v1";

export class ApiUnreachableError extends Error {
  constructor(cause?: unknown) {
    super("Could not reach the pave-api backend");
    this.name = "ApiUnreachableError";
    this.cause = cause;
  }
}

export class ApiResponseError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiResponseError";
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      // pave-api authenticates mutating requests via a session cookie
      // (see docs/adr/0002-pave-api-authentication.md) - "include" is
      // required for that cookie to be sent cross-origin to the API's port.
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    });
  } catch (cause) {
    throw new ApiUnreachableError(cause);
  }

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const body = await response.json();
      if (body?.error) message = body.error;
    } catch {
      // ignore body parse failures, keep default message
    }
    throw new ApiResponseError(response.status, message);
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

/* ────────────────────────────── Authentication ──────────────────────────── */
// pave-api authenticates via GitHub OAuth and a session cookie - see
// docs/adr/0002-pave-api-authentication.md. It may also be running with
// PAVE_API_DISABLE_AUTH=true (local dev), in which case /auth/me 404s
// (auth routes aren't registered) and every mutating call just works
// unauthenticated - callers here treat that the same as "not signed in"
// for display purposes, since either way there's no identity to show.

const API_ORIGIN = API_BASE_URL.replace(/\/api\/v1\/?$/, "");

export type CurrentUser = {
  login: string;
  teams: string[];
};

export async function getCurrentUser(): Promise<CurrentUser | null> {
  try {
    const response = await fetch(`${API_ORIGIN}/auth/me`, { credentials: "include" });
    if (!response.ok) return null;
    const body = (await response.json()) as unknown;
    if (!body || typeof body !== "object" || typeof (body as CurrentUser).login !== "string") {
      return null;
    }
    return body as CurrentUser;
  } catch {
    return null;
  }
}

/** Navigates the browser into the GitHub OAuth flow; returns here on success. */
export function loginUrl(): string {
  return `${API_ORIGIN}/auth/github/login`;
}

export async function logout(): Promise<void> {
  try {
    await fetch(`${API_ORIGIN}/auth/logout`, { method: "POST", credentials: "include" });
  } catch {
    // best-effort - the cookie is httpOnly so there's nothing else to clear client-side
  }
}

export async function checkHealth(): Promise<boolean> {
  try {
    const base = API_BASE_URL.replace(/\/api\/v1\/?$/, "");
    const response = await fetch(`${base}/healthz`);
    return response.ok;
  } catch {
    return false;
  }
}

/* ────────────────────────────── Create service ──────────────────────────── */

export type Runtime = "go" | "node" | "python";
export type Exposure = "internal" | "public";
export type Tier = "tier-1" | "tier-2" | "tier-3";

export type CreateServiceRequest = {
  name: string;
  team: string;
  runtime: Runtime;
  exposure: Exposure;
  database: boolean;
  tier: Tier;
};

export type CreateServiceResponse = {
  jobId: string;
  statusUrl: string;
};

export function createService(payload: CreateServiceRequest): Promise<CreateServiceResponse> {
  return request<CreateServiceResponse>("/services", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export type JobStatus =
  | "queued"
  | "validating"
  | "scaffolding"
  | "writing_manifests"
  | "opening_pr"
  | "completed"
  | "failed";

export type JobStep = {
  name: string;
  state: "pending" | "active" | "done" | "failed";
  message?: string;
  timestamp?: string;
};

export type Job = {
  jobId: string;
  status: JobStatus;
  steps: JobStep[];
  prUrl?: string;
  dryRun: boolean;
  error?: string;
};

export function getJob(jobId: string): Promise<Job> {
  return request<Job>(`/jobs/${encodeURIComponent(jobId)}`);
}

/* ────────────────────────────── Cost estimate ──────────────────────────── */

export type CostEstimate = {
  monthlyUsdLow: number;
  monthlyUsdHigh: number;
  currency: "USD";
  breakdown: { item: string; monthlyUsd: number }[];
  disclaimer: string;
};

export function getCostEstimate(params: {
  tier: string;
  exposure: string;
  database: boolean;
}): Promise<CostEstimate> {
  const query = new URLSearchParams({
    tier: params.tier,
    exposure: params.exposure,
    database: String(params.database),
  });
  return request<CostEstimate>(`/cost-estimate?${query.toString()}`);
}

/* ────────────────────────────── Access requests ──────────────────────────── */

export type AccessLevel = "read" | "write" | "admin";
export type AccessRequestStatus = "pending" | "approved" | "denied";

export type AccessRequest = {
  id: string;
  requester: string;
  team: string;
  namespace: string;
  level: AccessLevel;
  reason: string;
  status: AccessRequestStatus;
  approver?: string;
  note?: string;
  createdAt: string;
};

export type CreateAccessRequestPayload = {
  requester: string;
  team: string;
  namespace: string;
  level: AccessLevel;
  reason: string;
};

export function listAccessRequests(): Promise<AccessRequest[]> {
  return request<AccessRequest[]>("/access-requests");
}

export function createAccessRequest(payload: CreateAccessRequestPayload): Promise<AccessRequest> {
  return request<AccessRequest>("/access-requests", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

/* ────────────────────────────── Live services (optional) ──────────────────────────── */
// The static /catalog.json build artifact remains the canonical source for
// the Overview/Scorecards views. This is exposed for completeness against
// the contract but intentionally unused by those views today.

export function listServicesLive(): Promise<unknown[]> {
  return request<unknown[]>("/services");
}

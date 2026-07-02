import React, { useEffect, useState } from "react";
import {
  ApiResponseError,
  ApiUnreachableError,
  createAccessRequest,
  listAccessRequests,
  type AccessLevel,
  type AccessRequest,
} from "../../lib/api";
import { EmptyState, InlineError, Skeleton } from "../components";
import { IconKey } from "../icons";

function describeError(err: unknown): string {
  if (err instanceof ApiUnreachableError) {
    return "Backend unreachable — start pave-api locally (see docs) to submit or view access requests for real.";
  }
  if (err instanceof ApiResponseError) {
    if (err.status === 401)
      return "Sign in required — use the account menu above to sign in with GitHub.";
    if (err.status === 403)
      return "Your GitHub team doesn't have approval rights for this request.";
    return `pave-api rejected the request: ${err.message}`;
  }
  return err instanceof Error ? err.message : "Unknown error";
}

function StatusPill({ status }: { status: AccessRequest["status"] }) {
  const cls =
    status === "approved"
      ? "badge-success"
      : status === "denied"
        ? "badge-danger"
        : "badge-warning";
  return <span className={`badge ${cls}`}>{status}</span>;
}

const ACCESS_LEVELS: { value: AccessLevel; label: string; description: string }[] = [
  { value: "read", label: "Read", description: "View resources, logs, and configuration." },
  {
    value: "write",
    label: "Write",
    description: "Deploy and modify resources within the namespace.",
  },
  { value: "admin", label: "Admin", description: "Manage RBAC and namespace-level policy." },
];

export function RequestAccess() {
  const [requester, setRequester] = useState("");
  const [team, setTeam] = useState("");
  const [namespace, setNamespace] = useState("");
  const [level, setLevel] = useState<AccessLevel>("read");
  const [reason, setReason] = useState("");

  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState<AccessRequest | null>(null);

  const [requests, setRequests] = useState<AccessRequest[] | null>(null);
  const [listLoading, setListLoading] = useState(true);
  const [listError, setListError] = useState<string | null>(null);

  function loadRequests() {
    setListLoading(true);
    setListError(null);
    listAccessRequests()
      .then(setRequests)
      .catch((err) => setListError(describeError(err)))
      .finally(() => setListLoading(false));
  }

  useEffect(() => {
    loadRequests();
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setSubmitError(null);
    try {
      const created = await createAccessRequest({ requester, team, namespace, level, reason });
      setSubmitted(created);
      setReason("");
      loadRequests();
    } catch (err) {
      setSubmitError(describeError(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div>
        <h1 className="text-xl font-semibold text-pave-text mb-1">Request access</h1>
        <p className="text-sm text-pave-text-muted mb-6">
          Submits to a human approver — access is <strong>not</strong> granted automatically.
        </p>

        {submitted && (
          <div className="mb-4 rounded-lg border border-pave-warning/30 bg-pave-warning/5 px-4 py-3 text-sm text-pave-warning">
            Request submitted and is <strong>pending approval</strong>. You'll see it update in the
            list once an approver acts on it.
          </div>
        )}

        <form onSubmit={handleSubmit} className="card p-5 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label
                htmlFor="access-requester"
                className="block text-sm font-medium text-pave-text mb-1"
              >
                Requester
              </label>
              <input
                id="access-requester"
                required
                value={requester}
                onChange={(e) => setRequester(e.target.value)}
                placeholder="you@pavestack.io"
                className="search-input"
              />
            </div>
            <div>
              <label
                htmlFor="access-team"
                className="block text-sm font-medium text-pave-text mb-1"
              >
                Team
              </label>
              <input
                id="access-team"
                required
                value={team}
                onChange={(e) => setTeam(e.target.value)}
                placeholder="team-payments"
                className="search-input"
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="access-namespace"
              className="block text-sm font-medium text-pave-text mb-1"
            >
              Namespace
            </label>
            <input
              id="access-namespace"
              required
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              placeholder="payments"
              className="search-input"
            />
          </div>

          <fieldset>
            <legend className="block text-sm font-medium text-pave-text mb-2">Access level</legend>
            <div className="space-y-2">
              {ACCESS_LEVELS.map((opt) => (
                <label
                  key={opt.value}
                  aria-label={opt.label}
                  className={`flex items-start gap-2 rounded-lg border px-3 py-2 text-sm cursor-pointer ${
                    level === opt.value
                      ? "border-pave-accent ring-1 ring-pave-accent"
                      : "border-pave-border hover:border-pave-border-strong"
                  }`}
                >
                  <input
                    type="radio"
                    name="level"
                    className="mt-0.5"
                    checked={level === opt.value}
                    onChange={() => setLevel(opt.value)}
                  />
                  <span>
                    <span className="font-medium text-pave-text">{opt.label}</span>
                    <span className="block text-xs text-pave-text-muted">{opt.description}</span>
                  </span>
                </label>
              ))}
            </div>
          </fieldset>

          <div>
            <label
              htmlFor="access-reason"
              className="block text-sm font-medium text-pave-text mb-1"
            >
              Reason
            </label>
            <textarea
              id="access-reason"
              required
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Why do you need this access?"
              className="search-input resize-none"
            />
          </div>

          {submitError && (
            <InlineError message={submitError} onRetry={handleSubmit as unknown as () => void} />
          )}

          <button type="submit" disabled={submitting} className="btn btn-primary w-full">
            {submitting ? "Submitting…" : "Submit request"}
          </button>
        </form>
      </div>

      <div>
        <h2 className="text-sm font-semibold text-pave-text uppercase tracking-wider mb-3">
          Existing requests
        </h2>
        {listLoading && (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        )}
        {!listLoading && listError && <InlineError message={listError} onRetry={loadRequests} />}
        {!listLoading && !listError && requests && requests.length === 0 && (
          <EmptyState
            icon={<IconKey />}
            title="No access requests yet"
            description="Requests submitted here will appear with their approval status."
          />
        )}
        {!listLoading && !listError && requests && requests.length > 0 && (
          <ul className="space-y-2">
            {requests.map((r) => (
              <li key={r.id} className="card p-3">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-sm font-medium text-pave-text">{r.namespace}</span>
                  <StatusPill status={r.status} />
                </div>
                <p className="text-xs text-pave-text-muted mt-1">
                  {r.requester} · {r.team} · <span className="uppercase">{r.level}</span>
                </p>
                <p className="text-xs text-pave-text-secondary mt-1">{r.reason}</p>
                {r.note && (
                  <p className="text-xs text-pave-text-muted mt-1 italic">
                    Approver note: {r.note}
                  </p>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

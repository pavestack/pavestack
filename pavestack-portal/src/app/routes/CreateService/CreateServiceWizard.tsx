import React, { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import {
  RUNTIME_OPTIONS,
  TIER_OPTIONS,
  WIZARD_STEPS,
  initialWizardState,
  isStepValid,
  validateName,
  validateTeam,
  type WizardState,
  type WizardStepIndex,
} from "../../../lib/wizard";
import {
  ApiResponseError,
  ApiUnreachableError,
  createService,
  getCostEstimate,
  getJob,
  type CostEstimate,
  type Job,
} from "../../../lib/api";
import { InlineError } from "../../components";
import { StepTracker } from "../../StepTracker";
import { IconCheck, IconExternalLink } from "../../icons";

function StepProgress({ current }: { current: WizardStepIndex }) {
  return (
    <ol className="flex items-center gap-2 mb-6" aria-label="Wizard progress">
      {WIZARD_STEPS.map((label, i) => {
        const state = i < current ? "done" : i === current ? "active" : "pending";
        return (
          <li key={label} className="flex items-center gap-2 flex-1">
            <div className="flex items-center gap-2 min-w-0">
              <span
                className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-semibold ${
                  state === "done"
                    ? "bg-pave-success text-pave-accent-contrast"
                    : state === "active"
                      ? "bg-pave-accent text-pave-accent-contrast"
                      : "bg-pave-surface-hover text-pave-text-muted border border-pave-border"
                }`}
                aria-current={state === "active" ? "step" : undefined}
              >
                {state === "done" ? <IconCheck className="w-3 h-3" /> : i + 1}
              </span>
              <span
                className={`text-xs font-medium truncate hidden sm:inline ${state === "pending" ? "text-pave-text-muted" : "text-pave-text"}`}
              >
                {label}
              </span>
            </div>
            {i < WIZARD_STEPS.length - 1 && <span className="flex-1 h-px bg-pave-border" />}
          </li>
        );
      })}
    </ol>
  );
}

export function CreateServiceWizard() {
  const [step, setStep] = useState<WizardStepIndex>(0);
  const [state, setState] = useState<WizardState>(initialWizardState);
  const [touched, setTouched] = useState(false);

  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [job, setJob] = useState<Job | null>(null);
  const [jobId, setJobId] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [cost, setCost] = useState<CostEstimate | null>(null);
  const [costError, setCostError] = useState<string | null>(null);
  const [costLoading, setCostLoading] = useState(false);

  useEffect(() => {
    if (step !== 3) return;
    setCostLoading(true);
    setCostError(null);
    getCostEstimate({ tier: state.tier, exposure: state.exposure, database: state.database })
      .then(setCost)
      .catch((err) => setCostError(describeError(err)))
      .finally(() => setCostLoading(false));
  }, [step, state.tier, state.exposure, state.database]);

  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  function describeError(err: unknown): string {
    if (err instanceof ApiUnreachableError) {
      return "Backend unreachable — start pave-api locally (see docs) to run this for real.";
    }
    if (err instanceof ApiResponseError) {
      return `pave-api rejected the request: ${err.message}`;
    }
    return err instanceof Error ? err.message : "Unknown error";
  }

  function update<K extends keyof WizardState>(key: K, value: WizardState[K]) {
    setState((s) => ({ ...s, [key]: value }));
  }

  function goNext() {
    setTouched(true);
    if (!isStepValid(step, state)) return;
    setTouched(false);
    setStep((s) => Math.min(s + 1, 3) as WizardStepIndex);
  }

  function goBack() {
    setTouched(false);
    setStep((s) => Math.max(s - 1, 0) as WizardStepIndex);
  }

  function pollJob(id: string) {
    pollRef.current = setInterval(async () => {
      try {
        const latest = await getJob(id);
        setJob(latest);
        if (latest.status === "completed" || latest.status === "failed") {
          if (pollRef.current) clearInterval(pollRef.current);
        }
      } catch (err) {
        setSubmitError(describeError(err));
        if (pollRef.current) clearInterval(pollRef.current);
      }
    }, 1500);
  }

  async function handleSubmit() {
    setSubmitting(true);
    setSubmitError(null);
    try {
      const res = await createService(state);
      setJobId(res.jobId);
      try {
        const first = await getJob(res.jobId);
        setJob(first);
        if (first.status !== "completed" && first.status !== "failed") {
          pollJob(res.jobId);
        }
      } catch (err) {
        setSubmitError(describeError(err));
      }
    } catch (err) {
      setSubmitError(describeError(err));
    } finally {
      setSubmitting(false);
    }
  }

  function retry() {
    setSubmitError(null);
    setJob(null);
    setJobId(null);
    handleSubmit();
  }

  if (jobId) {
    return (
      <div className="max-w-2xl">
        <h1 className="text-xl font-semibold text-pave-text mb-1">Creating {state.name}</h1>
        <p className="text-sm text-pave-text-muted mb-6">
          Job <span className="font-mono">{jobId}</span> — polling{" "}
          <code className="font-mono">GET /api/v1/jobs/{jobId}</code> every 1.5s.
        </p>

        {submitError && (
          <div className="mb-4">
            <InlineError message={submitError} onRetry={retry} />
          </div>
        )}

        {job && (
          <div className="card p-5">
            {job.dryRun && <div className="badge badge-neutral mb-4">dry run</div>}
            <StepTracker steps={job.steps} />
            {job.status === "completed" && job.prUrl && (
              <div className="mt-4 pt-4 border-t border-pave-border">
                <a href={job.prUrl} target="_blank" rel="noreferrer" className="btn btn-primary">
                  View pull request
                  <IconExternalLink className="w-3.5 h-3.5" />
                </a>
              </div>
            )}
            {job.status === "failed" && job.error && (
              <div className="mt-4">
                <InlineError message={job.error} onRetry={retry} />
              </div>
            )}
          </div>
        )}

        <Link to="/" className="inline-block mt-6 text-sm text-pave-accent hover:underline">
          ← Back to catalog
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-xl font-semibold text-pave-text mb-1">Create a new service</h1>
      <p className="text-sm text-pave-text-muted mb-6">
        Mirrors <code className="font-mono">pave create-service</code> — scaffolds the golden-path
        template and opens a PR.
      </p>

      <StepProgress current={step} />

      <div className="card p-5">
        {step === 0 && (
          <div className="space-y-4">
            <div>
              <label
                htmlFor="wizard-name"
                className="block text-sm font-medium text-pave-text mb-1"
              >
                Service name
              </label>
              <input
                id="wizard-name"
                type="text"
                value={state.name}
                onChange={(e) => update("name", e.target.value)}
                placeholder="payments"
                className="search-input"
              />
              {touched && validateName(state.name) && (
                <p className="text-xs text-pave-danger mt-1">{validateName(state.name)}</p>
              )}
            </div>
            <div>
              <label
                htmlFor="wizard-team"
                className="block text-sm font-medium text-pave-text mb-1"
              >
                Owning team
              </label>
              <input
                id="wizard-team"
                type="text"
                value={state.team}
                onChange={(e) => update("team", e.target.value)}
                placeholder="team-payments"
                className="search-input"
              />
              {touched && validateTeam(state.team) && (
                <p className="text-xs text-pave-danger mt-1">{validateTeam(state.team)}</p>
              )}
            </div>
          </div>
        )}

        {step === 1 && (
          <fieldset className="space-y-3">
            <legend className="text-sm font-medium text-pave-text mb-1">Runtime</legend>
            {RUNTIME_OPTIONS.map((opt) => (
              <label
                key={opt.value}
                title={opt.note}
                className={`flex items-center gap-3 rounded-lg border px-3 py-2.5 text-sm ${
                  opt.available
                    ? "border-pave-border cursor-pointer hover:border-pave-border-strong"
                    : "border-pave-border opacity-50 cursor-not-allowed"
                } ${state.runtime === opt.value ? "ring-1 ring-pave-accent border-pave-accent" : ""}`}
              >
                <input
                  type="radio"
                  name="runtime"
                  value={opt.value}
                  checked={state.runtime === opt.value}
                  disabled={!opt.available}
                  onChange={() => update("runtime", opt.value)}
                />
                <span className="font-medium text-pave-text">{opt.label}</span>
                {!opt.available && <span className="badge badge-neutral ml-auto">Coming soon</span>}
              </label>
            ))}
            <p className="text-xs text-pave-text-muted">
              Only Go is scaffolded by the golden-path template today; Node.js and Python runtimes
              are planned.
            </p>
          </fieldset>
        )}

        {step === 2 && (
          <div className="space-y-5">
            <fieldset>
              <legend className="text-sm font-medium text-pave-text mb-2">Exposure</legend>
              <div className="flex gap-3">
                {(["internal", "public"] as const).map((exp) => (
                  <label
                    key={exp}
                    className={`flex-1 text-center rounded-lg border px-3 py-2 text-sm cursor-pointer capitalize ${
                      state.exposure === exp
                        ? "border-pave-accent ring-1 ring-pave-accent text-pave-accent"
                        : "border-pave-border text-pave-text-secondary hover:border-pave-border-strong"
                    }`}
                  >
                    <input
                      type="radio"
                      name="exposure"
                      className="sr-only"
                      checked={state.exposure === exp}
                      onChange={() => update("exposure", exp)}
                    />
                    {exp}
                  </label>
                ))}
              </div>
            </fieldset>

            <label className="flex items-center gap-2 text-sm text-pave-text">
              <input
                type="checkbox"
                checked={state.database}
                onChange={(e) => update("database", e.target.checked)}
              />
              Requires a managed database
            </label>

            <fieldset>
              <legend className="text-sm font-medium text-pave-text mb-2">Tier</legend>
              <div className="space-y-2">
                {TIER_OPTIONS.map((opt) => (
                  <label
                    key={opt.value}
                    className={`block rounded-lg border px-3 py-2.5 text-sm cursor-pointer ${
                      state.tier === opt.value
                        ? "border-pave-accent ring-1 ring-pave-accent"
                        : "border-pave-border hover:border-pave-border-strong"
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <input
                        type="radio"
                        name="tier"
                        checked={state.tier === opt.value}
                        onChange={() => update("tier", opt.value)}
                      />
                      <span className="font-medium text-pave-text">{opt.label}</span>
                    </div>
                    <p className="text-xs text-pave-text-muted mt-1 ml-6">{opt.description}</p>
                    <p className="text-xs text-pave-text-muted ml-6">
                      {opt.replicas} · {opt.sizing}
                    </p>
                  </label>
                ))}
              </div>
            </fieldset>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-5">
            <div>
              <h2 className="text-sm font-medium text-pave-text mb-2">Review</h2>
              <dl className="grid grid-cols-2 gap-y-2 text-sm">
                <dt className="text-pave-text-muted">Name</dt>
                <dd className="text-pave-text font-mono">{state.name}</dd>
                <dt className="text-pave-text-muted">Team</dt>
                <dd className="text-pave-text">{state.team}</dd>
                <dt className="text-pave-text-muted">Runtime</dt>
                <dd className="text-pave-text">{state.runtime}</dd>
                <dt className="text-pave-text-muted">Exposure</dt>
                <dd className="text-pave-text capitalize">{state.exposure}</dd>
                <dt className="text-pave-text-muted">Database</dt>
                <dd className="text-pave-text">{state.database ? "Yes" : "No"}</dd>
                <dt className="text-pave-text-muted">Tier</dt>
                <dd className="text-pave-text">{state.tier}</dd>
              </dl>
            </div>

            <div>
              <h2 className="text-sm font-medium text-pave-text mb-2">Estimated monthly cost</h2>
              {costLoading && <p className="text-sm text-pave-text-muted">Estimating…</p>}
              {costError && <InlineError message={costError} />}
              {cost && !costError && (
                <div>
                  <p className="text-xl font-bold tabular-nums text-pave-text">
                    ${cost.monthlyUsdLow}–${cost.monthlyUsdHigh}{" "}
                    <span className="text-sm font-normal text-pave-text-muted">/mo</span>
                  </p>
                  <p className="text-xs text-pave-text-muted mt-1">{cost.disclaimer}</p>
                </div>
              )}
            </div>

            {submitError && <InlineError message={submitError} onRetry={handleSubmit} />}
          </div>
        )}
      </div>

      <div className="flex items-center justify-between mt-4">
        <button type="button" onClick={goBack} disabled={step === 0} className="btn btn-secondary">
          Back
        </button>
        {step < 3 ? (
          <button type="button" onClick={goNext} className="btn btn-primary">
            Continue
          </button>
        ) : (
          <button
            type="button"
            onClick={handleSubmit}
            disabled={submitting}
            className="btn btn-primary"
          >
            {submitting ? "Submitting…" : "Create service"}
          </button>
        )}
      </div>
    </div>
  );
}

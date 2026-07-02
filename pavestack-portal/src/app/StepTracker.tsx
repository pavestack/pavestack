import React from "react";
import type { JobStep } from "../lib/api";
import { IconCheck, IconSpinner, IconX } from "./icons";

export function StepTracker({ steps }: { steps: JobStep[] }) {
  return (
    <ol className="space-y-0">
      {steps.map((step, i) => {
        const isLast = i === steps.length - 1;
        return (
          <li key={step.name} className="relative flex gap-3 pb-6 last:pb-0">
            {!isLast && (
              <span
                aria-hidden="true"
                className={`absolute left-[11px] top-6 bottom-0 w-px ${
                  step.state === "done" ? "bg-pave-success" : "bg-pave-border"
                }`}
              />
            )}
            <span
              className={`relative z-10 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs ${
                step.state === "done"
                  ? "bg-pave-success/15 border-pave-success text-pave-success"
                  : step.state === "active"
                    ? "bg-pave-accent/15 border-pave-accent text-pave-accent"
                    : step.state === "failed"
                      ? "bg-pave-danger/15 border-pave-danger text-pave-danger"
                      : "bg-pave-surface border-pave-border text-pave-text-muted"
              }`}
            >
              {step.state === "done" && <IconCheck className="w-3 h-3" />}
              {step.state === "active" && <IconSpinner className="w-3 h-3" />}
              {step.state === "failed" && <IconX className="w-3 h-3" />}
            </span>
            <div className="pt-0.5">
              <p
                className={`text-sm font-medium ${
                  step.state === "pending" ? "text-pave-text-muted" : "text-pave-text"
                }`}
              >
                {step.name}
              </p>
              {step.message && <p className="text-xs text-pave-text-muted mt-0.5">{step.message}</p>}
              {step.timestamp && (
                <p className="text-[11px] text-pave-text-muted font-mono tabular-nums mt-0.5">
                  {new Date(step.timestamp).toLocaleTimeString()}
                </p>
              )}
            </div>
          </li>
        );
      })}
    </ol>
  );
}

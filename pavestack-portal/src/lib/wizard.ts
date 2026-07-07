import type { Exposure, Runtime, Tier } from "./api";

export type WizardState = {
  name: string;
  team: string;
  runtime: Runtime;
  exposure: Exposure;
  database: boolean;
  tier: Tier;
};

export const initialWizardState: WizardState = {
  name: "",
  team: "",
  runtime: "go",
  exposure: "internal",
  database: false,
  tier: "tier-2",
};

export const WIZARD_STEPS = ["Name & team", "Runtime", "Shape & tier", "Review & submit"] as const;
export type WizardStepIndex = 0 | 1 | 2 | 3;

const NAME_PATTERN = /^[a-z][a-z0-9-]{1,48}[a-z0-9]$/;
const TEAM_PATTERN = /^[a-z][a-z0-9-]{1,62}[a-z0-9]$/;

export function validateName(name: string): string | null {
  if (!name) return "Service name is required.";
  if (!NAME_PATTERN.test(name)) {
    return "Use 3–50 lowercase letters, numbers, or hyphens, starting with a letter.";
  }
  return null;
}

export function validateTeam(team: string): string | null {
  if (!team) return "Owning team is required.";
  if (!TEAM_PATTERN.test(team)) {
    return "Use 3–64 lowercase letters, numbers, or hyphens, starting with a letter.";
  }
  return null;
}

/** Runtimes actually supported by the golden-path scaffold today. */
export const RUNTIME_OPTIONS: {
  value: Runtime;
  label: string;
  available: boolean;
  note?: string;
}[] = [
  { value: "go", label: "Go", available: true },
  {
    value: "node",
    label: "Node.js",
    available: false,
    note: "Coming soon — golden-path template is Go-only today",
  },
  {
    value: "python",
    label: "Python",
    available: false,
    note: "Coming soon — golden-path template is Go-only today",
  },
];

export const TIER_OPTIONS: {
  value: Tier;
  label: string;
  description: string;
  replicas: string;
  sizing: string;
}[] = [
  {
    value: "tier-1",
    label: "Tier 1 — Critical",
    description: "Customer-facing or revenue-critical. Paged on-call, highest headroom.",
    replicas: "3+ replicas, multi-AZ",
    sizing: "1 vCPU / 2Gi requests per pod",
  },
  {
    value: "tier-2",
    label: "Tier 2 — Standard",
    description: "Default for most internal services.",
    replicas: "2 replicas",
    sizing: "0.5 vCPU / 1Gi requests per pod",
  },
  {
    value: "tier-3",
    label: "Tier 3 — Low traffic",
    description: "Internal tools, batch jobs, low-traffic APIs.",
    replicas: "1 replica",
    sizing: "0.25 vCPU / 512Mi requests per pod",
  },
];

export function isStepValid(step: WizardStepIndex, state: WizardState): boolean {
  switch (step) {
    case 0:
      return validateName(state.name) === null && validateTeam(state.team) === null;
    case 1:
      return RUNTIME_OPTIONS.some((r) => r.value === state.runtime && r.available);
    case 2:
      return Boolean(state.exposure) && Boolean(state.tier);
    case 3:
      return true;
    default:
      return false;
  }
}

export function canAdvance(step: WizardStepIndex, state: WizardState): boolean {
  return isStepValid(step, state);
}

export const JOB_STEP_ORDER = [
  "validating",
  "scaffolding",
  "writing_manifests",
  "opening_pr",
  "completed",
] as const;

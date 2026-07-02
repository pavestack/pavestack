import { describe, test, expect } from "vitest";
import {
  canAdvance,
  initialWizardState,
  isStepValid,
  RUNTIME_OPTIONS,
  TIER_OPTIONS,
  validateName,
  validateTeam,
  WIZARD_STEPS,
  type WizardState,
} from "./wizard";

describe("wizard validation", () => {
  describe("validateName", () => {
    test("rejects empty names", () => {
      expect(validateName("")).toMatch(/required/);
    });

    test("rejects names that are too short, uppercase, or start with a digit", () => {
      expect(validateName("ab")).not.toBeNull();
      expect(validateName("Payments")).not.toBeNull();
      expect(validateName("1payments")).not.toBeNull();
    });

    test("accepts a valid DNS-safe name", () => {
      expect(validateName("payments-api")).toBeNull();
    });
  });

  describe("validateTeam", () => {
    test("rejects empty and malformed team slugs", () => {
      expect(validateTeam("")).toMatch(/required/);
      expect(validateTeam("Team_Payments")).not.toBeNull();
    });

    test("accepts a valid team slug", () => {
      expect(validateTeam("team-payments")).toBeNull();
    });
  });

  describe("RUNTIME_OPTIONS", () => {
    test("only go is available today; node/python are marked coming soon", () => {
      const go = RUNTIME_OPTIONS.find((r) => r.value === "go")!;
      const node = RUNTIME_OPTIONS.find((r) => r.value === "node")!;
      const python = RUNTIME_OPTIONS.find((r) => r.value === "python")!;
      expect(go.available).toBe(true);
      expect(node.available).toBe(false);
      expect(python.available).toBe(false);
      expect(node.note).toMatch(/coming soon/i);
    });
  });

  describe("TIER_OPTIONS", () => {
    test("defines replica/sizing guidance for all three tiers", () => {
      expect(TIER_OPTIONS).toHaveLength(3);
      for (const tier of TIER_OPTIONS) {
        expect(tier.replicas).toBeTruthy();
        expect(tier.sizing).toBeTruthy();
      }
    });
  });

  describe("isStepValid / canAdvance", () => {
    test("step 0 requires a valid name and team", () => {
      const state: WizardState = { ...initialWizardState, name: "", team: "" };
      expect(isStepValid(0, state)).toBe(false);
      expect(isStepValid(0, { ...state, name: "payments", team: "team-payments" })).toBe(true);
    });

    test("step 1 requires an available runtime", () => {
      expect(isStepValid(1, { ...initialWizardState, runtime: "go" })).toBe(true);
      expect(isStepValid(1, { ...initialWizardState, runtime: "node" })).toBe(false);
    });

    test("step 2 requires exposure and tier", () => {
      expect(isStepValid(2, initialWizardState)).toBe(true);
    });

    test("step 3 (review) is always valid to render", () => {
      expect(isStepValid(3, initialWizardState)).toBe(true);
    });

    test("canAdvance mirrors isStepValid", () => {
      expect(canAdvance(1, { ...initialWizardState, runtime: "python" })).toBe(false);
    });
  });

  test("WIZARD_STEPS has four labeled steps", () => {
    expect(WIZARD_STEPS).toHaveLength(4);
  });
});

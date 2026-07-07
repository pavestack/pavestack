package cost_test

import (
	"testing"

	"github.com/pavestack/pave/internal/cost"
)

func TestResolveTierDefaultsUnknown(t *testing.T) {
	if got := cost.ResolveTier(""); got != cost.DefaultTier {
		t.Errorf("expected default tier for empty input, got %s", got)
	}
	if got := cost.ResolveTier("not-a-tier"); got != cost.DefaultTier {
		t.Errorf("expected default tier for unknown input, got %s", got)
	}
	if got := cost.ResolveTier("tier-1"); got != cost.TierCritical {
		t.Errorf("expected tier-1 to resolve unchanged, got %s", got)
	}
}

func TestProfileForReplicaCounts(t *testing.T) {
	cases := []struct {
		tier       string
		wantDev    int
		wantProd   int
		wantReqCPU string
	}{
		{"tier-1", 2, 3, "250m"},
		{"tier-2", 1, 2, "100m"},
		{"tier-3", 1, 1, "50m"},
		{"", 1, 2, "100m"}, // defaults to tier-2
	}

	for _, tc := range cases {
		p := cost.ProfileFor(tc.tier)
		if p.Replicas.Dev != tc.wantDev || p.Replicas.Prod != tc.wantProd {
			t.Errorf("tier %q: expected replicas dev=%d prod=%d, got dev=%d prod=%d", tc.tier, tc.wantDev, tc.wantProd, p.Replicas.Dev, p.Replicas.Prod)
		}
		if p.Resources.RequestCPU != tc.wantReqCPU {
			t.Errorf("tier %q: expected request cpu %s, got %s", tc.tier, tc.wantReqCPU, p.Resources.RequestCPU)
		}
	}
}

func TestEstimateAddsDatabaseAndPublicLineItems(t *testing.T) {
	base := cost.Estimate(cost.TierStandard, "internal", false)
	if len(base.Breakdown) != 1 {
		t.Fatalf("expected 1 line item for internal/no-db, got %d", len(base.Breakdown))
	}

	withExtras := cost.Estimate(cost.TierStandard, "public", true)
	if len(withExtras.Breakdown) != 3 {
		t.Fatalf("expected 3 line items for public+database, got %d", len(withExtras.Breakdown))
	}
	if withExtras.MonthlyUSDLow <= base.MonthlyUSDLow {
		t.Error("expected public+database estimate to exceed the base estimate")
	}
	if withExtras.MonthlyUSDHigh <= withExtras.MonthlyUSDLow {
		t.Error("expected high estimate to exceed low estimate")
	}
}

func TestEstimateUnknownTierFallsBackToDefault(t *testing.T) {
	got := cost.Estimate(cost.Tier("bogus"), "internal", false)
	want := cost.Estimate(cost.DefaultTier, "internal", false)
	if got.MonthlyUSDLow != want.MonthlyUSDLow || got.MonthlyUSDHigh != want.MonthlyUSDHigh {
		t.Errorf("expected unknown tier to behave like default tier")
	}
}

// Package cost provides a static, transparent monthly cost estimate for a
// service request, and the resource/replica sizing profile associated with
// each service tier. It intentionally does not call any cloud pricing API —
// the numbers are a documented approximation so a developer sees an order of
// magnitude estimate before creating a service, not a bill.
package cost

import "fmt"

// Tier identifies the sizing/reliability profile a service is provisioned at.
type Tier string

const (
	TierCritical    Tier = "tier-1" // customer-facing / high traffic
	TierStandard    Tier = "tier-2" // default for most internal services
	TierLowTraffic  Tier = "tier-3" // internal tools, low QPS
	DefaultTier          = TierStandard
	DefaultRuntime       = "go"
	DefaultExposure      = "internal"
)

// ResourceProfile is the per-container Kubernetes resource sizing for a tier.
type ResourceProfile struct {
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

// ReplicaProfile is the per-environment replica count for a tier.
type ReplicaProfile struct {
	Dev  int
	Prod int
}

// Profile bundles the sizing and baseline compute cost range for a tier.
type Profile struct {
	Tier                Tier
	Label               string
	Resources           ResourceProfile
	Replicas            ReplicaProfile
	ComputeMonthlyLow   float64
	ComputeMonthlyHigh  float64
	DatabaseMonthlyLow  float64
	DatabaseMonthlyHigh float64
	PublicMonthlyLow    float64
	PublicMonthlyHigh   float64
}

var profiles = map[Tier]Profile{
	TierCritical: {
		Tier:  TierCritical,
		Label: "Tier 1 - critical / high traffic",
		Resources: ResourceProfile{
			RequestCPU: "250m", RequestMemory: "256Mi",
			LimitCPU: "1000m", LimitMemory: "1Gi",
		},
		Replicas:            ReplicaProfile{Dev: 2, Prod: 3},
		ComputeMonthlyLow:   95,
		ComputeMonthlyHigh:  150,
		DatabaseMonthlyLow:  60,
		DatabaseMonthlyHigh: 90,
		PublicMonthlyLow:    35, // ALB + WAF
		PublicMonthlyHigh:   55,
	},
	TierStandard: {
		Tier:  TierStandard,
		Label: "Tier 2 - standard internal service",
		Resources: ResourceProfile{
			RequestCPU: "100m", RequestMemory: "128Mi",
			LimitCPU: "500m", LimitMemory: "512Mi",
		},
		Replicas:            ReplicaProfile{Dev: 1, Prod: 2},
		ComputeMonthlyLow:   40,
		ComputeMonthlyHigh:  65,
		DatabaseMonthlyLow:  25,
		DatabaseMonthlyHigh: 45,
		PublicMonthlyLow:    15, // ALB target group + data transfer
		PublicMonthlyHigh:   25,
	},
	TierLowTraffic: {
		Tier:  TierLowTraffic,
		Label: "Tier 3 - low-traffic / internal tool",
		Resources: ResourceProfile{
			RequestCPU: "50m", RequestMemory: "64Mi",
			LimitCPU: "250m", LimitMemory: "256Mi",
		},
		Replicas:            ReplicaProfile{Dev: 1, Prod: 1},
		ComputeMonthlyLow:   18,
		ComputeMonthlyHigh:  30,
		DatabaseMonthlyLow:  15,
		DatabaseMonthlyHigh: 25,
		PublicMonthlyLow:    15,
		PublicMonthlyHigh:   25,
	},
}

// ResolveTier returns tier if it is a known tier, otherwise DefaultTier.
// Callers should treat this as the single place tier defaulting happens so
// the CLI, the pave-api backend, and gitops rendering never disagree about
// what an unset tier means.
func ResolveTier(tier string) Tier {
	t := Tier(tier)
	if _, ok := profiles[t]; ok {
		return t
	}
	return DefaultTier
}

// ProfileFor returns the sizing profile for a tier, defaulting unknown/empty
// tiers to DefaultTier.
func ProfileFor(tier string) Profile {
	return profiles[ResolveTier(tier)]
}

// LineItem is a single named contributor to the estimate.
type LineItem struct {
	Item          string
	MonthlyUSDLow float64
	MonthlyUSD    float64 // high end; kept alongside Low to render a range per item
}

// Result is the full monthly cost estimate for a service request.
type Result struct {
	Tier           Tier
	MonthlyUSDLow  float64
	MonthlyUSDHigh float64
	Currency       string
	Breakdown      []LineItem
	Disclaimer     string
}

// Estimate computes a rough monthly cost range for the given tier, exposure
// ("public" adds an ALB/ingress + data-transfer line), and database flag.
// This is a static approximation shared by `pave create-service` and the
// pave-api `/cost-estimate` endpoint - it is not a cloud billing API call.
func Estimate(tier Tier, exposure string, database bool) Result {
	p := ProfileFor(string(tier))

	est := Result{
		Tier:     p.Tier,
		Currency: "USD",
		Disclaimer: "Rough order-of-magnitude estimate based on shared-cluster " +
			"compute, storage, and networking assumptions for this tier. Not a " +
			"quote - actual AWS cost depends on real traffic, existing headroom " +
			"in the cluster, and data-transfer volume.",
	}

	est.Breakdown = append(est.Breakdown, LineItem{
		Item:          fmt.Sprintf("Compute (%s, %d dev + %d prod replicas)", p.Label, p.Replicas.Dev, p.Replicas.Prod),
		MonthlyUSDLow: p.ComputeMonthlyLow,
		MonthlyUSD:    p.ComputeMonthlyHigh,
	})
	est.MonthlyUSDLow += p.ComputeMonthlyLow
	est.MonthlyUSDHigh += p.ComputeMonthlyHigh

	if database {
		est.Breakdown = append(est.Breakdown, LineItem{
			Item:          "Managed database (RDS, tier-sized instance + backups)",
			MonthlyUSDLow: p.DatabaseMonthlyLow,
			MonthlyUSD:    p.DatabaseMonthlyHigh,
		})
		est.MonthlyUSDLow += p.DatabaseMonthlyLow
		est.MonthlyUSDHigh += p.DatabaseMonthlyHigh
	}

	if exposure == "public" {
		item := "Public ingress (ALB target group + data transfer)"
		if p.Tier == TierCritical {
			item = "Public ingress (ALB + WAF + data transfer)"
		}
		est.Breakdown = append(est.Breakdown, LineItem{
			Item:          item,
			MonthlyUSDLow: p.PublicMonthlyLow,
			MonthlyUSD:    p.PublicMonthlyHigh,
		})
		est.MonthlyUSDLow += p.PublicMonthlyLow
		est.MonthlyUSDHigh += p.PublicMonthlyHigh
	}

	return est
}

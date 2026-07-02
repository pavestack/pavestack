// Package apiserver implements the pave-api HTTP backend that powers the
// self-service portal's "Create service" and "Request access" flows. It
// reuses the same scaffold/gitops/validate packages the pave CLI uses, so
// a service created through the portal goes through the identical golden
// path as `pave create-service` — this package only adds an HTTP/async-job
// shell around that existing logic, it does not reimplement it.
package apiserver

import "time"

// Service is the catalog shape returned by GET /api/v1/services and
// GET /api/v1/services/:name. It mirrors, and is sourced from, the same
// catalog-info.yaml / scorecard.yaml files pavestack-portal's
// generate-catalog.mjs reads — this is intentionally the same source of
// truth, not a second data model.
type Service struct {
	Name         string                       `json:"name"`
	Team         string                       `json:"team"`
	Owner        string                       `json:"owner"`
	Lifecycle    string                       `json:"lifecycle"`
	System       string                       `json:"system"`
	Description  string                       `json:"description"`
	RepoURL      string                       `json:"repoUrl"`
	CreatedVia   string                       `json:"createdVia"` // "pave-cli" | "manual"
	Tier         string                       `json:"tier,omitempty"`
	Runtime      string                       `json:"runtime,omitempty"`
	Exposure     string                       `json:"exposure,omitempty"`
	Scorecard    Scorecard                    `json:"scorecard"`
	Environments map[string]EnvironmentStatus `json:"environments"`
}

type Scorecard struct {
	OverallScore int        `json:"overallScore"`
	Criteria     []Criteria `json:"criteria"`
}

type Criteria struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Weight   int    `json:"weight"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

// EnvironmentStatus is deliberately labeled "simulated" by callers when it
// is not backed by a live Argo CD/Prometheus read (see decisions.md) — this
// struct itself is agnostic to that; today's implementation reads the real
// image tag from platform-config but synthesizes sync/health, matching the
// same honesty constraint documented for pavestack-portal's catalog script.
type EnvironmentStatus struct {
	Status   string `json:"status"`
	Health   string `json:"health"`
	ImageTag string `json:"imageTag"`
}

// JobStatus enumerates the async create-service job lifecycle.
type JobStatus string

const (
	JobQueued           JobStatus = "queued"
	JobValidating       JobStatus = "validating"
	JobScaffolding      JobStatus = "scaffolding"
	JobWritingManifests JobStatus = "writing_manifests"
	JobOpeningPR        JobStatus = "opening_pr"
	JobCompleted        JobStatus = "completed"
	JobFailed           JobStatus = "failed"
)

type JobStep struct {
	Name      string    `json:"name"`
	State     string    `json:"state"` // "pending" | "running" | "done" | "failed"
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Job struct {
	JobID     string    `json:"jobId"`
	Status    JobStatus `json:"status"`
	Steps     []JobStep `json:"steps"`
	PRUrl     string    `json:"prUrl,omitempty"`
	DryRun    bool      `json:"dryRun"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateServiceRequest is the POST /api/v1/services request body.
type CreateServiceRequest struct {
	Name     string `json:"name"`
	Team     string `json:"team"`
	Runtime  string `json:"runtime"`
	Exposure string `json:"exposure"`
	Database bool   `json:"database"`
	Tier     string `json:"tier"`
}

// AccessRequest models a namespace/permission access request that requires
// explicit approval — the platform never silently auto-grants access.
type AccessRequest struct {
	ID        string    `json:"id"`
	Requester string    `json:"requester"`
	Team      string    `json:"team"`
	Namespace string    `json:"namespace"`
	Level     string    `json:"level"` // "read" | "write" | "admin"
	Reason    string    `json:"reason"`
	Status    string    `json:"status"` // "pending" | "approved" | "denied"
	Approver  string    `json:"approver,omitempty"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CostEstimateResponse is the GET /api/v1/cost-estimate response body.
type CostEstimateResponse struct {
	MonthlyUSDLow  float64        `json:"monthlyUsdLow"`
	MonthlyUSDHigh float64        `json:"monthlyUsdHigh"`
	Currency       string         `json:"currency"`
	Breakdown      []CostLineItem `json:"breakdown"`
	Disclaimer     string         `json:"disclaimer"`
}

type CostLineItem struct {
	Item          string  `json:"item"`
	MonthlyUSDLow float64 `json:"monthlyUsdLow"`
	MonthlyUSD    float64 `json:"monthlyUsd"`
}

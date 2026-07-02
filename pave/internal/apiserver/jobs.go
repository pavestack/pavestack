package apiserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/pavestack/pave/internal/gitops"
	"github.com/pavestack/pave/internal/scaffold"
	"github.com/pavestack/pave/internal/validate"
	"github.com/spf13/afero"
)

// JobStore runs create-service requests asynchronously through the exact
// same scaffold/gitops/validate calls pave create-service makes, and lets
// the portal poll for progress instead of blocking an HTTP request for the
// whole scaffold+PR flow.
type JobStore struct {
	mu       sync.RWMutex
	jobs     map[string]*Job
	repoRoot string
	dryRun   bool
}

func NewJobStore(repoRoot string, dryRun bool) *JobStore {
	return &JobStore{
		jobs:     make(map[string]*Job),
		repoRoot: repoRoot,
		dryRun:   dryRun,
	}
}

func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "job_" + hex.EncodeToString(b)
}

// Submit validates the request synchronously (so obviously-bad requests get
// an immediate 400) and, if valid, starts the scaffold+manifests+PR sequence
// in a goroutine, returning the job immediately in "queued" state.
func (s *JobStore) Submit(req CreateServiceRequest) (*Job, error) {
	sr := validate.ServiceRequest{
		Name:     req.Name,
		Team:     req.Team,
		Database: req.Database,
		Runtime:  req.Runtime,
		Exposure: req.Exposure,
		Tier:     req.Tier,
	}
	sr.ApplyDefaults()

	if err := validate.ValidateServiceRequest(s.repoRoot, sr); err != nil {
		return nil, err
	}

	job := &Job{
		JobID:     newJobID(),
		Status:    JobQueued,
		DryRun:    s.dryRun,
		CreatedAt: time.Now(),
		Steps: []JobStep{
			{Name: "validate", State: "done", Timestamp: time.Now()},
			{Name: "scaffold", State: "pending"},
			{Name: "write_manifests", State: "pending"},
			{Name: "open_pr", State: "pending"},
		},
	}

	s.mu.Lock()
	s.jobs[job.JobID] = job
	s.mu.Unlock()

	go s.run(job.JobID, sr)

	return job, nil
}

func (s *JobStore) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

func (s *JobStore) run(jobID string, sr validate.ServiceRequest) {
	s.setStatus(jobID, JobScaffolding)
	s.setStep(jobID, "scaffold", "running", "")

	fs := afero.NewOsFs()
	serviceDir, err := scaffold.CreateService(fs, s.repoRoot, sr)
	if err != nil {
		s.fail(jobID, "scaffold", fmt.Sprintf("scaffold failed: %v", err))
		return
	}
	s.setStep(jobID, "scaffold", "done", "Created "+serviceDir)

	s.setStatus(jobID, JobWritingManifests)
	s.setStep(jobID, "write_manifests", "running", "")
	if err := gitops.WriteTenantManifests(s.repoRoot, sr, serviceDir); err != nil {
		s.fail(jobID, "write_manifests", fmt.Sprintf("writing GitOps manifests failed: %v", err))
		return
	}
	s.setStep(jobID, "write_manifests", "done", fmt.Sprintf("Wrote platform-config/tenants/%s", sr.Name))

	s.setStatus(jobID, JobOpeningPR)
	s.setStep(jobID, "open_pr", "running", "")

	if s.dryRun {
		s.setStep(jobID, "open_pr", "done", "Dry-run: PR creation skipped (PAVE_API_DRY_RUN=true)")
		s.complete(jobID, "")
		return
	}

	prURL, err := gitops.CreatePullRequestURL(s.repoRoot, sr, "")
	if err != nil {
		// Scaffold + manifests already landed on disk even if the PR step
		// fails (e.g. git/gh unavailable) - report partial success rather
		// than JobFailed, matching what `pave create-service` itself does
		// (it warns and continues rather than erroring the whole command).
		s.setStep(jobID, "open_pr", "failed", fmt.Sprintf("PR creation skipped: %v", err))
		s.complete(jobID, "")
		return
	}
	s.setStep(jobID, "open_pr", "done", "Pull request opened")
	s.complete(jobID, prURL)
}

func (s *JobStore) setStatus(id string, status JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = status
	}
}

func (s *JobStore) setStep(id, name, state, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return
	}
	for i := range j.Steps {
		if j.Steps[i].Name == name {
			j.Steps[i].State = state
			j.Steps[i].Message = message
			j.Steps[i].Timestamp = time.Now()
		}
	}
}

func (s *JobStore) fail(id, step, message string) {
	s.setStep(id, step, "failed", message)
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = JobFailed
		j.Error = message
	}
}

func (s *JobStore) complete(id, prURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		j.Status = JobCompleted
		j.PRUrl = prURL
	}
}

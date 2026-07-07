package apiserver

import (
	"fmt"
	"testing"
)

// White-box tests for eviction, which operates on internal state directly
// rather than through the full Submit -> scaffold/validate flow.

func TestJobStoreEvictsOldestCompletedJobsOverLimit(t *testing.T) {
	s := NewJobStore("/tmp", true)

	for i := 0; i < maxTrackedJobs+10; i++ {
		id := fmt.Sprintf("job_%04d", i)
		s.jobs[id] = &Job{JobID: id, Status: JobCompleted}
		s.order = append(s.order, id)
	}

	s.mu.Lock()
	s.evictOldestCompletedLocked()
	s.mu.Unlock()

	if len(s.jobs) != maxTrackedJobs {
		t.Fatalf("expected %d jobs after eviction, got %d", maxTrackedJobs, len(s.jobs))
	}
	if _, ok := s.jobs["job_0000"]; ok {
		t.Error("expected oldest job to be evicted")
	}
	newest := fmt.Sprintf("job_%04d", maxTrackedJobs+9)
	if _, ok := s.jobs[newest]; !ok {
		t.Error("expected newest job to survive eviction")
	}
}

func TestJobStoreNeverEvictsRunningJobs(t *testing.T) {
	s := NewJobStore("/tmp", true)

	for i := 0; i < maxTrackedJobs+10; i++ {
		id := fmt.Sprintf("job_%04d", i)
		s.jobs[id] = &Job{JobID: id, Status: JobScaffolding}
		s.order = append(s.order, id)
	}

	s.mu.Lock()
	s.evictOldestCompletedLocked()
	s.mu.Unlock()

	if len(s.jobs) != maxTrackedJobs+10 {
		t.Fatalf("expected no eviction of still-running jobs, got %d jobs remaining", len(s.jobs))
	}
}

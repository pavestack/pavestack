package apiserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLogAppendsHashChain(t *testing.T) {
	dir := t.TempDir()
	l, err := newAuditLog(dir)
	if err != nil {
		t.Fatalf("newAuditLog: %v", err)
	}

	if err := l.append("create", "ar_1", "alice", "on-call"); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := l.append("approve", "ar_1", "platform-lead", ""); err != nil {
		t.Fatalf("append: %v", err)
	}

	// Reopening must replay and verify the chain without error, and pick
	// up where the sequence/hash left off.
	reopened, err := newAuditLog(dir)
	if err != nil {
		t.Fatalf("newAuditLog (reopen): %v", err)
	}
	if reopened.lastHash != l.lastHash {
		t.Errorf("expected reopened log to recover the same lastHash, got %q want %q", reopened.lastHash, l.lastHash)
	}
	if reopened.nextSeq != 2 {
		t.Errorf("expected nextSeq=2 after two entries, got %d", reopened.nextSeq)
	}
}

func TestAuditLogDetectsTamperedEntry(t *testing.T) {
	dir := t.TempDir()
	l, err := newAuditLog(dir)
	if err != nil {
		t.Fatalf("newAuditLog: %v", err)
	}
	if err := l.append("create", "ar_1", "alice", "on-call"); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := l.append("approve", "ar_1", "platform-lead", ""); err != nil {
		t.Fatalf("append: %v", err)
	}

	path := filepath.Join(dir, "access-requests.audit.ndjson")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), `"actor":"platform-lead"`, `"actor":"mallory"`, 1)
	if tampered == string(raw) {
		t.Fatal("test setup: tampering did not change the file")
	}
	if err := os.WriteFile(path, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := newAuditLog(dir); err == nil {
		t.Fatal("expected newAuditLog to detect the tampered entry and fail")
	}
}

func TestAuditLogDetectsTruncatedChain(t *testing.T) {
	dir := t.TempDir()
	l, err := newAuditLog(dir)
	if err != nil {
		t.Fatalf("newAuditLog: %v", err)
	}
	if err := l.append("create", "ar_1", "alice", "on-call"); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := l.append("approve", "ar_1", "platform-lead", ""); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := l.append("create", "ar_2", "bob", "rotation"); err != nil {
		t.Fatalf("append: %v", err)
	}

	path := filepath.Join(dir, "access-requests.audit.ndjson")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 audit lines, got %d", len(lines))
	}
	// Drop the middle entry - later entries' prev_hash no longer matches.
	withoutMiddle := lines[0] + "\n" + lines[2] + "\n"
	if err := os.WriteFile(path, []byte(withoutMiddle), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := newAuditLog(dir); err == nil {
		t.Fatal("expected newAuditLog to detect the broken hash chain and fail")
	}
}

func TestAccessRequestStoreWritesAuditTrail(t *testing.T) {
	dir := t.TempDir()
	store, err := NewAccessRequestStore(dir)
	if err != nil {
		t.Fatalf("NewAccessRequestStore: %v", err)
	}

	created, err := store.Create(AccessRequest{Requester: "alice", Team: "team-payments", Namespace: "payments", Level: "write", Reason: "on-call"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Decide(created.ID, "approve", "platform-lead", "looks fine"); err != nil {
		t.Fatalf("Decide: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "access-requests.audit.ndjson"))
	if err != nil {
		t.Fatalf("expected an audit log file to exist: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 audit entries (create + approve), got %d: %s", len(lines), raw)
	}
	if !strings.Contains(lines[0], `"actor":"alice"`) {
		t.Errorf("expected create entry to record the requester as actor, got %s", lines[0])
	}
	if !strings.Contains(lines[1], `"actor":"platform-lead"`) {
		t.Errorf("expected approve entry to record the approver as actor, got %s", lines[1])
	}

	// Reopening the store must succeed - proves the chain it just wrote is
	// self-consistent, not just that append() ran without error.
	if _, err := NewAccessRequestStore(dir); err != nil {
		t.Fatalf("expected store to reopen cleanly, got %v", err)
	}
}

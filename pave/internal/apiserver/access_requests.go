package apiserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// AccessRequestStore persists access requests to a JSON file so a restart
// of pave-api doesn't lose pending approvals. It intentionally never
// auto-approves - Create always lands in "pending", only Decide can move a
// request to "approved"/"denied", and Decide requires an approver identity.
//
// access-requests.json is the rebuildable "current state" view; every
// Create/Decide is additionally appended to audit, a hash-chained NDJSON
// log that exists specifically so "who approved what, and was this record
// altered after the fact" has an answer independent of that mutable file.
type AccessRequestStore struct {
	mu    sync.Mutex
	path  string
	data  map[string]AccessRequest
	audit *auditLog
}

func NewAccessRequestStore(runtimeDir string) (*AccessRequestStore, error) {
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return nil, err
	}
	audit, err := newAuditLog(runtimeDir)
	if err != nil {
		return nil, err
	}
	s := &AccessRequestStore{
		path:  filepath.Join(runtimeDir, "access-requests.json"),
		data:  map[string]AccessRequest{},
		audit: audit,
	}
	if raw, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(raw, &s.data)
	}
	return s, nil
}

func (s *AccessRequestStore) List() []AccessRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]AccessRequest, 0, len(s.data))
	for _, v := range s.data {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *AccessRequestStore) Create(req AccessRequest) (AccessRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	req.ID = "ar_" + hex.EncodeToString(b)
	req.Status = "pending"
	req.CreatedAt = time.Now()

	s.data[req.ID] = req
	if err := s.persistLocked(); err != nil {
		return AccessRequest{}, err
	}
	if err := s.audit.append("create", req.ID, req.Requester, req.Reason); err != nil {
		return AccessRequest{}, fmt.Errorf("persisted but failed to audit-log the create: %w", err)
	}
	return req, nil
}

func (s *AccessRequestStore) Decide(id, action, approver, note string) (AccessRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.data[id]
	if !ok {
		return AccessRequest{}, fmt.Errorf("access request %q not found", id)
	}
	switch action {
	case "approve":
		req.Status = "approved"
	case "deny":
		req.Status = "denied"
	default:
		return AccessRequest{}, fmt.Errorf("unknown action %q, expected approve or deny", action)
	}
	req.Approver = approver
	req.Note = note
	s.data[id] = req
	if err := s.persistLocked(); err != nil {
		return AccessRequest{}, err
	}
	if err := s.audit.append(action, id, approver, note); err != nil {
		return AccessRequest{}, fmt.Errorf("persisted but failed to audit-log the %s: %w", action, err)
	}
	return req, nil
}

func (s *AccessRequestStore) persistLocked() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}

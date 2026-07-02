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
type AccessRequestStore struct {
	mu   sync.Mutex
	path string
	data map[string]AccessRequest
}

func NewAccessRequestStore(runtimeDir string) (*AccessRequestStore, error) {
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return nil, err
	}
	s := &AccessRequestStore{
		path: filepath.Join(runtimeDir, "access-requests.json"),
		data: map[string]AccessRequest{},
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
	return req, s.persistLocked()
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
	return req, s.persistLocked()
}

func (s *AccessRequestStore) persistLocked() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}

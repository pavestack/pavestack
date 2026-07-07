package apiserver

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// genesisHash is the prev_hash of the first entry in a fresh audit log.
const genesisHash = "genesis"

// auditEntry is one append-only, hash-chained record of a sensitive
// access-request action. Hash covers every other field plus PrevHash, so
// altering or deleting a past entry (or reordering the file) breaks the
// chain from that point forward - detectable by loadTail without needing
// a separate signing key, since pave-api is both writer and verifier here.
type auditEntry struct {
	Seq       int64     `json:"seq"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "create" | "approve" | "deny"
	RequestID string    `json:"request_id"`
	// Actor is the requester (for "create") or the approver (for
	// "approve"/"deny") - the caller in access_requests.go/server.go
	// passes the verified session identity here when auth is enabled, so
	// the audit trail inherits that guarantee automatically.
	Actor    string `json:"actor"`
	Note     string `json:"note,omitempty"`
	PrevHash string `json:"prev_hash"`
	Hash     string `json:"hash"`
}

func (e auditEntry) computeHash() string {
	e.Hash = ""
	raw, _ := json.Marshal(e)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// auditLog is an append-only NDJSON file at
// <runtimeDir>/access-requests.audit.ndjson, separate from
// access-requests.json (which stays the rebuildable, mutable "current
// state" view - the audit log is what you'd hand an auditor asking "prove
// nothing here was altered after the fact").
type auditLog struct {
	mu       sync.Mutex
	path     string
	lastHash string
	nextSeq  int64
}

func newAuditLog(runtimeDir string) (*auditLog, error) {
	l := &auditLog{
		path:     filepath.Join(runtimeDir, "access-requests.audit.ndjson"),
		lastHash: genesisHash,
	}
	if err := l.loadTail(); err != nil {
		return nil, err
	}
	return l, nil
}

// loadTail replays the whole log to verify the hash chain and recover
// lastHash/nextSeq. Returns an error - refusing to start pave-api - if
// the chain is broken, since a broken chain means the log was altered
// out-of-band; that's exactly the condition this feature exists to catch,
// so silently continuing would defeat the point.
func (l *auditLog) loadTail() error {
	f, err := os.Open(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)

	prevHash := genesisHash
	var seq int64
	seen := false
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e auditEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return fmt.Errorf("audit log %s is corrupt: %w", l.path, err)
		}
		if e.PrevHash != prevHash {
			return fmt.Errorf("audit log %s hash chain broken at seq %d: expected prev_hash %q, got %q - the log may have been altered", l.path, e.Seq, prevHash, e.PrevHash)
		}
		if e.computeHash() != e.Hash {
			return fmt.Errorf("audit log %s entry %d hash mismatch - the log may have been altered", l.path, e.Seq)
		}
		prevHash = e.Hash
		seq = e.Seq
		seen = true
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	l.lastHash = prevHash
	if seen {
		l.nextSeq = seq + 1
	}
	return nil
}

func (l *auditLog) append(action, requestID, actor, note string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := auditEntry{
		Seq:       l.nextSeq,
		Timestamp: time.Now().UTC(),
		Action:    action,
		RequestID: requestID,
		Actor:     actor,
		Note:      note,
		PrevHash:  l.lastHash,
	}
	e.Hash = e.computeHash()

	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(raw); err != nil {
		return err
	}

	l.lastHash = e.Hash
	l.nextSeq++
	return nil
}

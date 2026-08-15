package timings

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Version is the current store format version.
	Version = 1
	// dateLayout is the LastSeen format.
	dateLayout = "2006-01-02"
)

// Store is the canonical timing store: one JSON document.
type Store struct {
	Version int                  `json:"version"`
	Tests   map[string]TestEntry `json:"tests"`
}

// TestEntry holds a ring buffer of the most recent durations.
type TestEntry struct {
	DurationsMs []int64 `json:"durations_ms"`
	LastSeen    string  `json:"last_seen"`
}

// Load reads the store from a local path or http(s) URL. A missing local
// file is a cold start: an empty store, no error. Empty ref = empty store.
func Load(ref string) (*Store, error) {
	if ref == "" {
		return &Store{Version: Version, Tests: map[string]TestEntry{}}, nil
	}
	data, err := readRef(ref)
	if err != nil {
		return nil, err
	}
	s := &Store{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parse timings %s: %w", ref, err)
	}
	if s.Version > Version {
		return nil, fmt.Errorf("timings %s has version %d, this build supports up to %d", ref, s.Version, Version)
	}
	if s.Tests == nil {
		s.Tests = map[string]TestEntry{}
	}
	return s, nil
}

func readRef(ref string) ([]byte, error) {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		resp, err := http.Get(ref)
		if err != nil {
			return nil, fmt.Errorf("fetch timings %s: %w", ref, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch timings %s: %s", ref, resp.Status)
		}
		return io.ReadAll(resp.Body)
	}
	data, err := os.ReadFile(ref)
	if os.IsNotExist(err) {
		return []byte(`{"version":1,"tests":{}}`), nil // cold start
	}
	if err != nil {
		return nil, fmt.Errorf("read timings %s: %w", ref, err)
	}
	return data, nil
}

// Merge folds a fresh duration (ms) per test id into the store, keeping the
// most recent `history` durations and stamping LastSeen with today. Returns
// the number of existing entries updated and the number of new ids added.
func (s *Store) Merge(durations map[string]int64, today time.Time, history int) (merged, added int) {
	if history < 1 {
		history = 1
	}
	date := today.Format(dateLayout)
	for id, d := range durations {
		e, ok := s.Tests[id]
		if !ok {
			e = TestEntry{}
			added++
		} else {
			merged++
		}
		e.DurationsMs = append(e.DurationsMs, d)
		if len(e.DurationsMs) > history {
			e.DurationsMs = e.DurationsMs[len(e.DurationsMs)-history:]
		}
		e.LastSeen = date
		s.Tests[id] = e
	}
	return merged, added
}

// Prune drops entries whose LastSeen is older than afterDays, returning the
// number pruned. Only report should call this — never a filtered discover:
// a unit-only split run must not delete integration timings.
func (s *Store) Prune(today time.Time, afterDays int) int {
	cutoff := today.AddDate(0, 0, -afterDays).Format(dateLayout)
	pruned := 0
	for id, e := range s.Tests {
		if e.LastSeen != "" && e.LastSeen < cutoff {
			delete(s.Tests, id)
			pruned++
		}
	}
	return pruned
}

// Save writes the store atomically (temp file + rename in the destination
// directory). The path must be local — remote writes are not supported.
func (s *Store) Save(path string) error {
	if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return fmt.Errorf("cannot write timings to a URL (%s); use --timings-out with a local path", path)
	}
	s.Version = Version
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode timings: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp timings file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { tmp.Close(); os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp timings file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp timings file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename timings into place: %w", err)
	}
	return nil
}

// ExpectedFor returns the expected duration (ms) for a test id and whether
// the store knows it. history caps how many recent durations are considered.
func (s *Store) ExpectedFor(id string, history int) (int64, bool) {
	e, ok := s.Tests[id]
	if !ok || len(e.DurationsMs) == 0 {
		return 0, false
	}
	ds := e.DurationsMs
	if history < 1 {
		history = 1
	}
	if history < len(ds) {
		ds = ds[len(ds)-history:]
	}
	return ExpectedDuration(ds), true
}

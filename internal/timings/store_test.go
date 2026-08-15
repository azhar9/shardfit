package timings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var today = time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

func TestLoadMissingIsColdStart(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != Version || len(s.Tests) != 0 {
		t.Fatalf("cold start = %+v, want version 1, empty tests", s)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "timings.json")
	s, _ := Load("")
	s.Merge(map[string]int64{"a": 120}, today, 5)
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	back, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Tests) != 1 || back.Tests["a"].DurationsMs[0] != 120 || back.Tests["a"].LastSeen != "2026-08-14" {
		t.Fatalf("roundtrip = %+v", back.Tests)
	}
	// no temp files left behind (atomic write)
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestMergeCapsHistoryAndCounts(t *testing.T) {
	s, _ := Load("")
	for i := int64(1); i <= 7; i++ {
		merged, added := s.Merge(map[string]int64{"a": i}, today, 5)
		if i == 1 && added != 1 {
			t.Fatalf("first merge added = %d, want 1", added)
		}
		if i > 1 && merged != 1 {
			t.Fatalf("merge %d updated = %d, want 1", i, merged)
		}
	}
	ds := s.Tests["a"].DurationsMs
	if len(ds) != 5 || ds[0] != 3 || ds[4] != 7 {
		t.Fatalf("ring buffer = %v, want [3 4 5 6 7]", ds)
	}
}

func TestMergeHistoryAtLeastOne(t *testing.T) {
	s, _ := Load("")
	s.Merge(map[string]int64{"a": 1}, today, 0)
	if len(s.Tests["a"].DurationsMs) != 1 {
		t.Fatalf("history 0 should keep 1 duration, got %v", s.Tests["a"].DurationsMs)
	}
}

func TestPrune(t *testing.T) {
	s, _ := Load("")
	s.Merge(map[string]int64{"fresh": 1}, today, 5)
	s.Merge(map[string]int64{"stale": 1}, today.AddDate(0, 0, -40), 5)
	pruned := s.Prune(today, 30)
	if pruned != 1 {
		t.Fatalf("pruned = %d, want 1", pruned)
	}
	if _, ok := s.Tests["fresh"]; !ok {
		t.Fatal("fresh entry was pruned")
	}
	if _, ok := s.Tests["stale"]; ok {
		t.Fatal("stale entry survived")
	}
}

func TestLoadURL(t *testing.T) {
	body := `{"version":1,"tests":{"a":{"durations_ms":[5],"last_seen":"2026-08-14"}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	defer srv.Close()
	s, err := Load(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Tests) != 1 {
		t.Fatalf("url load = %+v", s.Tests)
	}
}

func TestLoadURLNotFound(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	if _, err := Load(srv.URL); err == nil {
		t.Fatal("want error for 404")
	}
}

func TestLoadRejectsNewerVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "timings.json")
	os.WriteFile(path, []byte(`{"version":2,"tests":{}}`), 0o644)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("err = %v, want version error", err)
	}
}

func TestSaveRejectsURL(t *testing.T) {
	s, _ := Load("")
	if err := s.Save("https://example.com/timings.json"); err == nil {
		t.Fatal("want error saving to URL")
	}
}

func TestSaveCleanupOnRenameFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "timings.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	s, _ := Load("")
	s.Merge(map[string]int64{"a": 1}, today, 5)
	if err := s.Save(path); err == nil {
		t.Fatal("want error when target path is a directory")
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestExpectedFor(t *testing.T) {
	s, _ := Load("")
	for i := int64(100); i <= 700; i += 100 {
		s.Merge(map[string]int64{"a": i}, today, 7)
	}
	got, ok := s.ExpectedFor("a", 5)
	if !ok {
		t.Fatal("ExpectedFor should know a")
	}
	// uses only the last 5 durations (300..700) → weighted median 600
	if got != 600 {
		t.Fatalf("ExpectedFor = %d, want 600", got)
	}
	if _, ok := s.ExpectedFor("nope", 5); ok {
		t.Fatal("ExpectedFor should not know nope")
	}
}

func TestStoreJSONShape(t *testing.T) {
	s, _ := Load("")
	s.Merge(map[string]int64{"a": 7}, today, 5)
	data, _ := json.Marshal(s)
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["version"] != float64(1) {
		t.Fatalf("version = %v", m["version"])
	}
}

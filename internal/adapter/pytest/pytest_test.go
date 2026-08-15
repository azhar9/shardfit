package pytest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter"
)

// pytest's junitxml classname is a dotted module path; ParseDurations must
// reconstruct collect-style ids against the working tree.
func TestParseDurationsReconstructsFile(t *testing.T) {
	dir := t.TempDir()
	mk := func(rel string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("tests/test_api.py")
	t.Chdir(dir) // not parallel-safe: no t.Parallel in this package

	xml := `<testsuites><testsuite name="pytest">
  <testcase classname="tests.test_api" name="test_create" time="1.234"/>
  <testcase classname="tests.test_api.TestClass" name="test_method" time="0.5"/>
  <testcase classname="missing.mod" name="test_gone" time="0.1"/>
</testsuite></testsuites>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{
		"tests/test_api.py::test_create":            1234,
		"tests/test_api.py::TestClass::test_method": 500,
		"missing/mod.py::test_gone":                 100,
	}
	for id, d := range want {
		if got[id] != d {
			t.Fatalf("got[%q] = %d, want %d (all: %v)", id, got[id], d, got)
		}
	}
}

func TestParseDurationsSumsRetries(t *testing.T) {
	xml := `<testsuite name="pytest">
  <testcase classname="mod" name="test_x" time="1"/>
  <testcase classname="mod" name="test_x" time="2"/>
</testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["mod.py::test_x"] != 3000 {
		t.Fatalf("got = %v", got)
	}
}

func TestDiscoverReadsInputList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	if err := os.WriteFile(path, []byte("tests/a.py::test_one\ntests/b.py::TestK::test_two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tests, err := New().Discover(adapter.DiscoverConfig{Input: path})
	if err != nil {
		t.Fatal(err)
	}
	if tests[0].File != "tests/a.py" || tests[1].File != "tests/b.py" {
		t.Fatalf("files = %q, %q", tests[0].File, tests[1].File)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityTest {
		t.Fatal("pytest should shard at test granularity")
	}
}

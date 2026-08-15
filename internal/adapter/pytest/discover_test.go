//go:build !windows

package pytest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter"
)

func TestDiscoverParsesCollectOutput(t *testing.T) {
	// fake pytest on PATH: emits two collect ids, a blank line, and the
	// summary line — the parse loop must keep the ids and skip the rest.
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf '%s\\n' 'tests/a.py::test_one' 'tests/b.py::TestK::test_two' '' '3 tests collected in 0.00s'\n"
	if err := os.WriteFile(filepath.Join(bin, "pytest"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	tests, err := New().Discover(adapter.DiscoverConfig{})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ id, file string }{
		{"tests/a.py::test_one", "tests/a.py"},
		{"tests/b.py::TestK::test_two", "tests/b.py"},
	}
	if len(tests) != len(want) {
		t.Fatalf("tests = %+v, want %d entries", tests, len(want))
	}
	for i, w := range want {
		if tests[i].ID != w.id || tests[i].File != w.file {
			t.Fatalf("tests[%d] = %+v, want ID=%q File=%q", i, tests[i], w.id, w.file)
		}
	}
}

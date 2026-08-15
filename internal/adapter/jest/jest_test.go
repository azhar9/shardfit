package jest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter"
)

func TestParseDurationsByFile(t *testing.T) {
	xml := `<testsuites><testsuite name="jest">
  <testcase classname="src/__tests__/api.test.js" name="api creates user" time="2.5"/>
  <testcase classname="src/__tests__/api.test.js" name="api lists users" time="1.5"/>
</testsuite></testsuites>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["src/__tests__/api.test.js"] != 4000 {
		t.Fatalf("got = %v, want 4000 summed by file", got)
	}
}

func TestParseDurationsNonPathClassname(t *testing.T) {
	xml := `<testsuite name="jest"><testcase classname="api" name="creates user" time="0.5"/></testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["api > creates user"] != 500 {
		t.Fatalf("got = %v", got)
	}
}

func TestDiscoverReadsInputList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	os.WriteFile(path, []byte("src/a.test.js\nsrc/b.test.js\n"), 0o644)
	tests, err := New().Discover(adapter.DiscoverConfig{Input: path})
	if err != nil {
		t.Fatal(err)
	}
	if tests[0].ID != "src/a.test.js" || tests[0].File != "src/a.test.js" {
		t.Fatalf("tests = %+v (File must equal ID at file granularity)", tests)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityFile {
		t.Fatal("jest must shard at file granularity")
	}
}

func TestDiscoverRelativizesPaths(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(dir, "src", "a.test.js")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n%%s\\n%%s\\n' '%s' '/outside/cwd/b.test.js' 'src/rel.test.js'\n", abs)
	if err := os.WriteFile(filepath.Join(bin, "jest"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Chdir(dir)

	tests, err := New().Discover(adapter.DiscoverConfig{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"src/a.test.js", "/outside/cwd/b.test.js", "src/rel.test.js"}
	if len(tests) != len(want) {
		t.Fatalf("tests = %+v, want %d entries", tests, len(want))
	}
	for i, w := range want {
		if tests[i].ID != w || tests[i].File != w {
			t.Fatalf("tests[%d] = %+v, want ID=File=%q", i, tests[i], w)
		}
	}
}

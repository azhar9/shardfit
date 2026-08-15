package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/azhar9/shardfit/internal/adapter/generic"
	"github.com/azhar9/shardfit/internal/timings"
)

func writeList(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "list.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGenericSplitEndToEnd(t *testing.T) {
	dir := t.TempDir()
	list := writeList(t, dir, "t1\nt2\nt3\nt4\n")
	cmd := newSplitCmd(generic.New())
	cmd.SetArgs([]string{"-n", "2", "--input", list, "--timings", filepath.Join(dir, "timings.json"), "--out-dir", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, name := range []string{"bucket-1.txt", "bucket-2.txt"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line != "" {
				all = append(all, line)
			}
		}
	}
	if len(all) != 4 {
		t.Fatalf("bucket files contain %d ids, want 4: %v", len(all), all)
	}
}

func TestSplitColdStartSucceeds(t *testing.T) {
	dir := t.TempDir()
	list := writeList(t, dir, "a\nb\nc\nd\n")
	cmd := newSplitCmd(generic.New())
	cmd.SetArgs([]string{"-n", "2", "--input", list, "--out-dir", dir}) // no --timings
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cold start must succeed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "bucket-1.txt")); err != nil {
		t.Fatal("bucket-1.txt not written")
	}
}

func TestSplitErrorsWhenStoreKnowsNothing(t *testing.T) {
	dir := t.TempDir()
	// store exists but holds an unrelated test: selection is 100% unknown
	store, _ := timings.Load("")
	store.Merge(map[string]int64{"other::test": 10}, time.Now(), 5)
	timingsPath := filepath.Join(dir, "timings.json")
	if err := store.Save(timingsPath); err != nil {
		t.Fatal(err)
	}
	list := writeList(t, dir, "a\nb\n")
	cmd := newSplitCmd(generic.New())
	cmd.SetArgs([]string{"-n", "2", "--input", list, "--timings", timingsPath, "--out-dir", dir})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no history") {
		t.Fatalf("err = %v, want unknown-percentage error", err)
	}
}

func TestSplitTooManyBuckets(t *testing.T) {
	dir := t.TempDir()
	list := writeList(t, dir, "a\nb\n")
	cmd := newSplitCmd(generic.New())
	cmd.SetArgs([]string{"-n", "3", "--input", list, "--out-dir", dir})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "more buckets") {
		t.Fatalf("err = %v, want 'more buckets' error", err)
	}
}

func TestSplitEstimateOnlyWritesNothing(t *testing.T) {
	dir := t.TempDir()
	list := writeList(t, dir, "a\nb\nc\nd\n")
	cmd := newSplitCmd(generic.New())
	cmd.SetArgs([]string{"-n", "2", "--input", list, "--out-dir", dir, "--estimate-only"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "bucket-1.txt")); !os.IsNotExist(err) {
		t.Fatal("--estimate-only must not write bucket files")
	}
}

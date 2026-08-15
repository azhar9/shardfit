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

const reportFixture = `<testsuites><testsuite name="s">
  <testcase classname="Suite" name="test_a" time="1.5"/>
  <testcase classname="Suite" name="test_b" time="0.5"/>
</testsuite></testsuites>`

func TestGenericReportEndToEnd(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "results.xml")
	if err := os.WriteFile(xmlPath, []byte(reportFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	timingsPath := filepath.Join(dir, "timings.json")
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", xmlPath, "--timings-out", timingsPath})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	store, err := timings.Load(timingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if store.Tests["Suite.test_a"].DurationsMs[0] != 1500 || store.Tests["Suite.test_b"].DurationsMs[0] != 500 {
		t.Fatalf("store = %+v", store.Tests)
	}
}

func TestReportMergesIntoExistingStore(t *testing.T) {
	dir := t.TempDir()
	store, _ := timings.Load("")
	store.Merge(map[string]int64{"Suite.test_a": 1000}, time.Now(), 5)
	timingsPath := filepath.Join(dir, "timings.json")
	if err := store.Save(timingsPath); err != nil {
		t.Fatal(err)
	}
	xmlPath := filepath.Join(dir, "results.xml")
	if err := os.WriteFile(xmlPath, []byte(reportFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", xmlPath, "--timings", timingsPath})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	back, _ := timings.Load(timingsPath)
	ds := back.Tests["Suite.test_a"].DurationsMs
	if len(ds) != 2 || ds[0] != 1000 || ds[1] != 1500 {
		t.Fatalf("durations = %v, want [1000 1500]", ds)
	}
}

func TestReportRequiresDestination(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "results.xml")
	os.WriteFile(xmlPath, []byte(reportFixture), 0o644)
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", xmlPath})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no destination") {
		t.Fatalf("err = %v, want 'no destination'", err)
	}
}

func TestReportRejectsURLDestination(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "results.xml")
	os.WriteFile(xmlPath, []byte(reportFixture), 0o644)
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", xmlPath, "--timings", "https://example.com/timings.json"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "URL") {
		t.Fatalf("err = %v, want URL rejection without a network call", err)
	}
}

func TestReportRejectsBadPruneAfter(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "results.xml")
	if err := os.WriteFile(xmlPath, []byte(reportFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	timingsPath := filepath.Join(dir, "timings.json")
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", xmlPath, "--timings-out", timingsPath, "--prune-after", "0"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "prune-after") {
		t.Fatalf("err = %v, want 'prune-after' rejection", err)
	}
	if _, err := os.Stat(timingsPath); !os.IsNotExist(err) {
		t.Fatalf("timings file was written despite bad --prune-after (stat err = %v)", err)
	}
}

func TestReportNoFiles(t *testing.T) {
	dir := t.TempDir()
	cmd := newReportCmd(generic.New())
	cmd.SetArgs([]string{"--junit-xml", filepath.Join(dir, "nope-*.xml"), "--timings-out", filepath.Join(dir, "t.json")})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "no JUnit XML") {
		t.Fatalf("err = %v, want 'no JUnit XML files'", err)
	}
}

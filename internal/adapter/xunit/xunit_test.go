package xunit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter"
)

func TestDiscoverRequiresInput(t *testing.T) {
	if _, err := New().Discover(adapter.DiscoverConfig{}); err == nil {
		t.Fatal("want error when no --input")
	}
}

func TestDiscoverReadsList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	os.WriteFile(path, []byte("MyApp.Tests.ApiTests.test_create\n"), 0o644)
	tests, err := New().Discover(adapter.DiscoverConfig{Input: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 1 || tests[0].ID != "MyApp.Tests.ApiTests.test_create" {
		t.Fatalf("tests = %+v", tests)
	}
}

func TestParseDurationsJUnitLogger(t *testing.T) {
	xml := `<testsuite name="MyApp.Tests.ApiTests" tests="2">
  <testcase classname="MyApp.Tests.ApiTests" name="test_create" time="1.2"/>
  <testcase classname="MyApp.Tests.ApiTests" name="test_read" time="0.4"/>
</testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["MyApp.Tests.ApiTests.test_create"] != 1200 || got["MyApp.Tests.ApiTests.test_read"] != 400 {
		t.Fatalf("got = %v", got)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityTest {
		t.Fatal("xunit should shard at test granularity")
	}
}

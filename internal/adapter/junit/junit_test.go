package junit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter"
)

func TestDiscoverRequiresInput(t *testing.T) {
	if _, err := New().Discover(adapter.DiscoverConfig{}); err == nil {
		t.Fatal("want error when no --input (Java discovery needs a build)")
	}
}

func TestDiscoverReadsList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	os.WriteFile(path, []byte("com.example.ApiTest.testCreate\n"), 0o644)
	tests, err := New().Discover(adapter.DiscoverConfig{Input: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 1 || tests[0].ID != "com.example.ApiTest.testCreate" {
		t.Fatalf("tests = %+v", tests)
	}
}

func TestParseDurationsSurefire(t *testing.T) {
	xml := `<testsuite name="com.example.ApiTest" tests="3">
  <testcase classname="com.example.ApiTest" name="testCreate" time="0.125"/>
  <testcase classname="com.example.ApiTest" name="testCreate" time="0.075"/>
  <testcase classname="com.example.ApiTest" name="testRead" time="0.3"/>
</testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["com.example.ApiTest.testCreate"] != 200 || got["com.example.ApiTest.testRead"] != 300 {
		t.Fatalf("got = %v", got)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityTest {
		t.Fatal("junit should shard at test granularity")
	}
}

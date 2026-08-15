package generic

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
	os.WriteFile(path, []byte("suite::test_a\nsuite::test_b\n"), 0o644)
	tests, err := New().Discover(adapter.DiscoverConfig{Input: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 2 || tests[0].ID != "suite::test_a" {
		t.Fatalf("tests = %+v", tests)
	}
}

func TestParseDurations(t *testing.T) {
	xml := `<testsuites><testsuite name="s">
  <testcase classname="Suite" name="test_a" time="1.5"/>
  <testcase classname="Suite" name="test_a" time="0.5"/>
  <testcase name="no_class" time="0.25"/>
</testsuite></testsuites>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["Suite.test_a"] != 2000 || got["no_class"] != 250 {
		t.Fatalf("got = %v", got)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityTest {
		t.Fatal("generic should shard at test granularity")
	}
}

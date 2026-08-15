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

func TestParseDurationsFullDisplayNames(t *testing.T) {
	// real JUnitTestLogger 1.1.0 shape: name already carries the full
	// display name; the classname prefix must not be duplicated
	xml := `<testsuite name="Demo.Tests.CalculatorTests" tests="2">
  <testcase classname="Demo.Tests.CalculatorTests" name="Demo.Tests.CalculatorTests.TestFastOne" time="0.3"/>
  <testcase classname="Demo.Tests.CalculatorTests" name="Demo.Tests.CalculatorTests.TestScaled(ms: 100)" time="0.1"/>
</testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{
		"Demo.Tests.CalculatorTests.TestFastOne":         300,
		"Demo.Tests.CalculatorTests.TestScaled(ms: 100)": 100,
	}
	for id, d := range want {
		if got[id] != d {
			t.Fatalf("got[%q] = %d, want %d (all: %v)", id, got[id], d, got)
		}
	}
}

func TestParseDurationsTheoryCases(t *testing.T) {
	// [Theory] + [InlineData] rows: JUnitTestLogger emits one testcase per
	// row, with the arguments in the display name — each row keeps its own
	// id and duration, exactly as `dotnet test --list-tests` lists them
	xml := `<testsuite name="MyApp.Tests.ApiTests" tests="3">
  <testcase classname="MyApp.Tests.ApiTests" name="Add_Numbers_ReturnsExpectedSum(a: 2, b: 3, expected: 5)" time="0.1"/>
  <testcase classname="MyApp.Tests.ApiTests" name="Add_Numbers_ReturnsExpectedSum(a: 10, b: 5, expected: 15)" time="0.15"/>
  <testcase classname="MyApp.Tests.ApiTests" name="Add_Numbers_ReturnsExpectedSum(a: -1, b: -1, expected: -2)" time="0.12"/>
</testsuite>`
	got, err := New().ParseDurations([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["MyApp.Tests.ApiTests.Add_Numbers_ReturnsExpectedSum(a: 2, b: 3, expected: 5)"] != 100 {
		t.Fatalf("theory row 1 = %v", got)
	}
	if len(got) != 3 {
		t.Fatalf("theory rows must stay distinct, got %d entries: %v", len(got), got)
	}
}

func TestGranularity(t *testing.T) {
	if New().Granularity() != adapter.GranularityTest {
		t.Fatal("xunit should shard at test granularity")
	}
}

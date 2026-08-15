package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fake struct{}

func (fake) Name() string             { return "fake" }
func (fake) Granularity() Granularity { return GranularityTest }
func (fake) Discover(DiscoverConfig) ([]Test, error) {
	return []Test{{ID: "a"}}, nil
}
func (fake) ParseDurations([]byte) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func TestRegisterGet(t *testing.T) {
	Register(fake{})
	a, err := Get("fake")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name() != "fake" {
		t.Fatalf("got %q", a.Name())
	}
	if !strings.Contains(Names(), "fake") {
		t.Fatalf("Names = %q", Names())
	}
}

func TestGetUnknown(t *testing.T) {
	_, err := Get("nope")
	if err == nil || !strings.Contains(err.Error(), "available") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadListFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	os.WriteFile(path, []byte("a\nb\n\na\n"), 0o644)
	tests, err := ReadList(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 3 || tests[0].ID != "a" || tests[1].ID != "b" || tests[2].ID != "a" {
		t.Fatalf("tests = %+v", tests)
	}
}

func TestReadListStdin(t *testing.T) {
	tests, err := ReadList("-", strings.NewReader("x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 1 || tests[0].ID != "x" {
		t.Fatalf("tests = %+v", tests)
	}
}

func TestReadListMissingFile(t *testing.T) {
	if _, err := ReadList(filepath.Join(t.TempDir(), "nope.txt"), nil); err == nil {
		t.Fatal("want error for missing list file")
	}
}

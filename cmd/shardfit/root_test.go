package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/azhar9/shardfit/internal/adapter/generic"
)

func TestRootHasAllAdapterTrees(t *testing.T) {
	root := newRootCmd()
	for _, name := range []string{"pytest", "jest", "junit", "xunit", "generic"} {
		for _, verb := range []string{"discover", "split", "report"} {
			cmd, _, err := root.Find([]string{name, verb})
			if err != nil || cmd == nil {
				t.Fatalf("missing %s %s: %v", name, verb, err)
			}
		}
	}
}

func TestDiscoverPrintsIds(t *testing.T) {
	dir := t.TempDir()
	list := writeList(t, dir, "a\nb\n")
	cmd := newDiscoverCmd(generic.New())
	cmd.SetArgs([]string{"--input", list})
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Split(strings.TrimSpace(out.String()), "\n"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRootVersion(t *testing.T) {
	root := newRootCmd()
	if root.Version == "" {
		t.Fatal("root should carry a version")
	}
}

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExpandXMLGlobDedupeAndLiteral(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"r1.xml", "r2.xml"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("<testsuite/>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// glob + duplicate via a second arg → deduped to the two files
	got, err := expandXML([]string{
		filepath.Join(dir, "r*.xml"),
		filepath.Join(dir, "r1.xml"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expandXML = %v, want exactly r1.xml and r2.xml", got)
	}
	names := map[string]bool{}
	for _, g := range got {
		names[filepath.Base(g)] = true
	}
	if !names["r1.xml"] || !names["r2.xml"] {
		t.Fatalf("expandXML = %v, missing a file", got)
	}

	// glob matching nothing contributes nothing
	got, err = expandXML([]string{filepath.Join(dir, "nope-*.xml")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("empty glob = %v, want []", got)
	}

	// literal paths are kept so their read error is clear
	got, err = expandXML([]string{filepath.Join(dir, "missing.xml")})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{filepath.Join(dir, "missing.xml")}) {
		t.Fatalf("literal = %v", got)
	}
}

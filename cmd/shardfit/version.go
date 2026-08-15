package main

import (
	"runtime/debug"
	"strings"
)

// version is stamped at build time by goreleaser
// (-X main.version={{.Version}}). go-installed builds fall back to the
// module version embedded by the Go toolchain (e.g. "0.1.0" for @v0.1.0
// installs); dev checkouts get a pseudo-version identifying the commit.
var version = "0.1.0-dev"

func init() {
	if version != "0.1.0-dev" {
		return // stamped by ldflags
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := strings.TrimPrefix(info.Main.Version, "v"); v != "" && v != "(devel)" {
			version = v
		}
	}
}

# Contributing

Thanks for helping out. shardfit is small on purpose: one binary, one JSON
format, adapters as thin wrappers. Contributions should stay in that shape.

## Setup

```bash
brew install go golangci-lint   # macOS; any recent Go on other systems
go build ./cmd/shardfit
go test ./...
```

## Workflow

1. Open an issue first for anything beyond a small fix.
2. Branch from `main`; commit with conventional-commit messages.
3. PR titles must be conventional (`feat:`, `fix:`, `docs:`, `test:`,
   `chore:`) — CI checks this.
4. `go test ./...`, `gofmt -l .`, and `golangci-lint run` must pass.

## Adding an adapter

Adapters implement one interface (`internal/adapter/adapter.go`):

```go
type Adapter interface {
    Name() string
    Discover(cfg DiscoverConfig) ([]Test, error)          // Test{ID, File}
    Granularity() Granularity                             // Test or File
    ParseDurations(junitXML []byte) (map[string]int64, error)
}
```

Steps:

1. Create `internal/adapter/<name>/<name>.go` implementing the interface.
   - `Discover`: run the framework's list command (filters passed through
     verbatim — never translated) or read `cfg.Input`.
   - `ParseDurations`: map JUnit XML testcases to the same ids `Discover`
     emits, summing retries (use `junitxml.SumByID`).
   - Ids must be stable across runs; document the format in your docs file.
2. Register it in `cmd/shardfit/main.go`.
3. Add `docs/adapters/<name>.md` (discover/split/report setup).
4. Add a fixture-XML test alongside the adapter — mirror the existing ones.

## Releases

Tag `vX.Y.Z` on main; goreleaser builds and attaches binaries automatically.

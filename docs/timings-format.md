# Timing store format

One JSON document, version 1:

```json
{
  "version": 1,
  "tests": {
    "tests/test_api.py::test_create[user]": {
      "durations_ms": [1234, 1190, 1400],
      "last_seen": "2026-08-13"
    }
  }
}
```

- `durations_ms`: ring buffer of the last N runs (N = `--history`, default 5).
- `last_seen`: date (`YYYY-MM-DD`) stamped by the last `report` that saw the
  test. Entries unseen for `--prune-after` days (default 30) are dropped by
  `report`.
- Keys are adapter test ids (see `docs/adapters/`). **Ids must be stable
  across runs** — if your discovery setup changes ids, timings silently
  reset. `report` prints how many new tests it saw, which is your signal.
- Unknown keys are ignored (forward compatibility). A file with a `version`
  newer than the binary supports is rejected with a clear error — never
  silently mis-parsed.

## Storage and CI

shardfit reads the store from a local path or http(s) URL and writes only to
local paths (atomic tmp+rename). Keep it in your CI's cache or artifacts
bucket, or a GCS/S3 path that your CI syncs. Cache it **per branch** and let
only the default branch's copy propagate — the standard CI-cache pattern —
so parallel branches don't race on writes. The report job is the only writer.

## Fast tests and overhead

Durations come from JUnit XML's per-test `time` attribute, in milliseconds —
sub-millisecond tests round to 0 or 1ms. Very fast suites are handled
correctly: when every test weighs ~0, the split spreads them evenly by count,
which is already the optimal balance for equal durations. Note that the
estimate models the test itself, not the wall-clock around it (interpreter
startup, imports, fixtures) — so for suites dominated by overhead, expect the
estimated per-bucket totals to understate real wall time while the split
itself stays balanced. Jest's file-level sharding captures more of that
overhead per unit.

## Hygiene rules

- Deleted tests: stop appearing in reports, then age out. Pruning happens
  only in `report` — never on filtered discovery — so a unit-only split run
  can't delete your integration timings.
- New tests: estimated as the median of known tests at split time; their
  first real duration lands at the next report.
- Cold start (missing file): split proceeds on estimates with a warning.

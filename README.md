# shardfit

[![CI](https://img.shields.io/github/actions/workflow/status/azhar9/shardfit/ci.yml?branch=main)](https://github.com/azhar9/shardfit/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/azhar9/shardfit/branch/main/graph/badge.svg)](https://codecov.io/gh/azhar9/shardfit)
[![License: MIT](https://img.shields.io/github/license/azhar9/shardfit)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/azhar9/shardfit)](go.mod)

Split your test suite into N buckets using past runtime data, so parallel CI
shards finish at roughly the same time. One static binary, any framework —
pytest, jest, JUnit, xunit, or anything that can emit a test list and JUnit
XML. No SaaS, no DB: your timing data is one JSON file you keep wherever you
want (CI cache, artifacts, GCS, S3).

## How it works

```
Job 1: split                    Job 2: test (matrix 1..8)         Job 3: report (needs: test)
  restore timings.json           download bucket-$N.txt            download results-*.xml
  shardfit pytest split -n 8    pytest $(cat bucket-$N.txt)        shardfit pytest report \
    --timings timings.json        --junitxml=results-$N.xml          --junit-xml "results-*.xml" \
  upload buckets                 upload results-$N.xml                --timings timings.json
                                                                     upload timings.json → cache
```

The split phase reads the timing store, estimates each test's duration, and
writes `bucket-1.txt … bucket-N.txt` (one test id per line). The report phase
folds the JUnit XML your runner already produces back into the store. Buckets
run tests completely normally — shardfit is never in the hot path.

## Install

Download the latest binary from
[Releases](https://github.com/azhar9/shardfit/releases), or build from source:

```bash
go install github.com/azhar9/shardfit/cmd/shardfit@latest
```

## Quickstart (pytest)

1. **Split** (one job, before your test matrix):

```bash
shardfit pytest split -n 8 --timings timings.json --out-dir buckets
# writes buckets/bucket-1.txt … bucket-8.txt
```

2. **Run** (matrix job `$SHARD`), with JUnit XML enabled:

```bash
pytest $(cat buckets/bucket-$SHARD.txt) --junitxml=results-$SHARD.xml
```

3. **Report** (one job, after the matrix):

```bash
shardfit pytest report --junit-xml "results-*.xml" --timings timings.json
```

First run is a cold start: buckets are balanced by test count, a warning says
so, and from the second run on the split is duration-aware.

## Adapters

| Adapter | discover | granularity | JUnit XML source |
|---|---|---|---|
| `pytest` | `pytest --collect-only -q` (+ `--filter`) | test | `--junitxml` |
| `jest` | `jest --listTests` (+ `--filter`) | file | jest-junit reporter |
| `junit` | `--input` list | test | surefire |
| `xunit` | `--input` list | test | JUnitTestLogger |
| `generic` | `--input` list | test | any |

See `docs/adapters/` for per-adapter setup and `docs/ci/` for CI wiring
(GitHub Actions, GitLab, CircleCI, Jenkins).

## How the split works

Each test's expected duration is a recency-weighted median of its last 5
runs, dampened when the history is flaky, capped at the suite's P99. Tests
with no history get the median of known tests. Buckets are filled
longest-first into the least-loaded bucket (LPT — provably within ~33% of
optimal). Deterministic for identical input.

## Docs

- [Timing store format](docs/timings-format.md)
- [Adapters](docs/adapters/)
- [CI integration](docs/ci/)
- [Contributing](CONTRIBUTING.md)

MIT licensed.

# pytest adapter

Shards at test granularity. Test ids look like
`tests/test_api.py::TestClass::test_name[param]` — exactly what
`pytest --collect-only -q` prints.

## discover

```bash
shardfit pytest discover              # all tests
shardfit pytest discover --filter "-k unit"   # filter forwarded verbatim
shardfit pytest discover --input list.txt     # skip pytest, read ids
```

## split

```bash
shardfit pytest split -n 8 --timings timings.json --out-dir buckets
```

`--group-by file` keeps same-file tests in one bucket (module-level fixtures).
The default groups at test granularity.

## report

Run pytest with `--junitxml` (built-in), then:

```bash
shardfit pytest report --junit-xml "results-*.xml" --timings timings.json
```

Run both commands from the repository root so file paths match. Durations
are reconstructed from pytest's dotted `classname` by checking candidate file
paths against the working tree, so the report job needs the checkout.

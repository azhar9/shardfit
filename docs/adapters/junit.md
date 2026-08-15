# junit adapter (Java)

Shards at test granularity. Test ids are surefire-style
`com.example.ApiTest.testMethod`. Java discovery needs a compiled build,
which shardfit doesn't drive — produce a test list from your build and pass
it with `--input`:

```bash
# e.g. from a surefire scan or your own script
shardfit junit split -n 8 --input tests.txt --timings timings.json
```

surefire already writes JUnit XML per module; point `report` at them:

```bash
shardfit junit report --junit-xml "target/surefire-reports/*.xml" --timings timings.json
```

Retried tests (duplicate testcases) are summed per id.

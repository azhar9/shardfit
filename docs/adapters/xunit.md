# xunit adapter (.NET)

Shards at test granularity. Test ids are `Namespace.Class.Method`. Produce
the list from the SDK and pass it with `--input`:

```bash
dotnet test --list-tests > tests.txt
shardfit xunit split -n 8 --input tests.txt --timings timings.json
```

For durations, add the JUnitTestLogger package so `dotnet test` emits JUnit
XML, then:

```bash
shardfit xunit report --junit-xml "results-*.xml" --timings timings.json
```

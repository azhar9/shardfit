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

Theory rows keep their arguments in the display name — each row is its own
test id with its own duration. Produce the list and run the tests with the
same SDK version: VSTest display-name formatting has changed across .NET
versions, and both sides must agree.

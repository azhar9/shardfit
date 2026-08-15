# generic adapter

The zero-code path for any framework: a test list in, JUnit XML for
durations. Ids are whatever your list contains — lines are opaque to
shardfit, so make them stable and unique.

```bash
shardfit generic split -n 8 --input tests.txt --timings timings.json
shardfit generic report --junit-xml "results-*.xml" --timings timings.json
```

Durations are matched by `classname.name` (or bare `name` when classname is
empty). If your runner's XML uses a different identity, prefer writing a
native adapter — see CONTRIBUTING.md.

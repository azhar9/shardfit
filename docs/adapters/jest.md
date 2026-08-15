# jest adapter

Shards at **file** granularity — jest runs each test file in its own process,
so files are never split across buckets. Test ids are repo-relative file
paths.

## discover

```bash
shardfit jest discover                      # jest --listTests under the hood
shardfit jest discover --filter "--testPathPattern=unit"   # verbatim passthrough
```

## split

```bash
shardfit jest split -n 8 --timings timings.json --out-dir buckets
```

## report

Add the `jest-junit` reporter (default `classNameTemplate` is the file path,
which is what the adapter expects):

```bash
npm i -D jest-junit
jest --reporters=default --reporters=jest-junit   # → junit.xml
shardfit jest report --junit-xml "results-*.xml" --timings timings.json
```

If you customize `classNameTemplate` to something without a path, the
adapter falls back to `classname > name` ids — keep the template stable.

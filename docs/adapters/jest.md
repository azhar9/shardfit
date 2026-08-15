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

Add the `jest-junit` reporter **with the file-path classNameTemplate**.
jest-junit v16+ defaults to the test title, which can never match
discovery ids — shardfit fails loudly without this setting:

```js
// jest.config.js
module.exports = {
  reporters: [
    "default",
    ["jest-junit", { classNameTemplate: "{filepath}" }],
  ],
};
```

```bash
npm i -D jest-junit
jest   # → junit.xml
shardfit jest report --junit-xml "results-*.xml" --timings timings.json
```

Custom `classNameTemplate` values without a path produce
`classname > name` ids that never match discovery ids — keep the template
stable and path-based.

Note: run from a non-symlinked checkout. macOS links `/tmp` to
`/private/tmp`; a checkout under a symlinked path leaves `--listTests`
output absolute, and those ids never join the timings store.

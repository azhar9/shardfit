# GitHub Actions

Split and report jobs plus a matrix in between. The store lives in the
workflow cache, keyed by branch so only each branch sees its own history
(merge to main refreshes the main key).

```yaml
jobs:
  split:
    runs-on: ubuntu-latest
    outputs:
      shards: ${{ steps.split.outputs.shards }}
    steps:
      - uses: actions/checkout@v4
      - name: Install shardfit
        run: |
          curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_Linux_x86_64.tar.gz | tar xz
          sudo mv shardfit /usr/local/bin/
      - uses: actions/cache@v4
        id: timings
        with:
          path: timings.json
          key: shardfit-timings-${{ github.ref_name }}-${{ github.sha }}
          restore-keys: |
            shardfit-timings-${{ github.ref_name }}-
            shardfit-timings-main-
      - id: split
        run: |
          shardfit pytest split -n 8 --timings timings.json --out-dir buckets
          echo "shards=$(ls buckets/bucket-*.txt | wc -l)" >> "$GITHUB_OUTPUT"
      - uses: actions/upload-artifact@v4
        with:
          name: buckets
          path: buckets/
      - uses: actions/upload-artifact@v4
        with:
          name: timings-read
          path: timings.json
          if-no-files-found: ignore

  test:
    needs: split
    runs-on: ubuntu-latest
    strategy:
      matrix:
        shard: [1, 2, 3, 4, 5, 6, 7, 8]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: buckets
      - run: pytest $(cat bucket-${{ matrix.shard }}.txt) --junitxml=results-${{ matrix.shard }}.xml
      - uses: actions/upload-artifact@v4
        with:
          name: results-${{ matrix.shard }}
          path: results-${{ matrix.shard }}.xml

  report:
    needs: test
    if: always()
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install shardfit
        run: |
          curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_Linux_x86_64.tar.gz | tar xz
          sudo mv shardfit /usr/local/bin/
      - uses: actions/download-artifact@v4
        with:
          pattern: results-*
          merge-multiple: true
      - uses: actions/cache@v4
        with:
          path: timings.json
          key: shardfit-timings-${{ github.ref_name }}-${{ github.sha }}
          restore-keys: |
            shardfit-timings-${{ github.ref_name }}-
            shardfit-timings-main-
      - run: shardfit pytest report --junit-xml "results-*.xml" --timings timings.json
```

Notes: `report` runs even when tests fail (`if: always()`), because failed
runs still carry real durations. `merge-multiple: true` unpacks every
`results-*` artifact into one directory.

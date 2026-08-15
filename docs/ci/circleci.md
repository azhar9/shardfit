# CircleCI

One workflow, three jobs; the store is a workspace artifact plus a
restore key falling back to the main branch's copy.

```yaml
version: 2.1

jobs:
  split:
    docker: [{image: cimg/base:stable}]
    steps:
      - checkout
      - restore_cache:
          keys:
            - shardfit-timings-{{ .Branch }}-{{ .Revision }}
            - shardfit-timings-{{ .Branch }}-
            - shardfit-timings-main-
      - run: |
          curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_Linux_x86_64.tar.gz | tar xz
          ./shardfit pytest split -n 8 --timings timings.json --out-dir buckets
      - persist_to_workspace:
          root: .
          paths: [buckets]

  test:
    docker: [{image: python:3.12}]
    parameters:
      shard: {type: integer}
    steps:
      - checkout
      - attach_workspace: {at: .}
      - run: pytest $(cat buckets/bucket-<< parameters.shard >>.txt) --junitxml=results-<< parameters.shard >>.xml
      - persist_to_workspace:
          root: .
          paths: [results-*.xml]

  report:
    docker: [{image: cimg/base:stable}]
    steps:
      - checkout
      - restore_cache:
          keys:
            - shardfit-timings-{{ .Branch }}-{{ .Revision }}
            - shardfit-timings-{{ .Branch }}-
            - shardfit-timings-main-
      - attach_workspace: {at: .}
      - run: |
          curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_Linux_x86_64.tar.gz | tar xz
          ./shardfit pytest report --junit-xml "results-*.xml" --timings timings.json
      - save_cache:
          key: shardfit-timings-{{ .Branch }}-{{ .Revision }}
          paths: [timings.json]

workflows:
  sharded:
    jobs:
      - split
      - test:
          requires: [split]
          matrix:
            parameters: {shard: [1, 2, 3, 4, 5, 6, 7, 8]}
      - report:
          requires: [test]
```

# GitLab CI

Same shape: split job, parallel matrix, report job. The store travels in the
job cache with a fallback key to the default branch.

```yaml
stages: [split, test, report]

.split-cache: &split-cache
  cache:
    - key: shardfit-timings-$CI_COMMIT_REF_SLUG
      paths: [timings.json]
      fallback_keys: [shardfit-timings-main]

split:
  stage: split
  <<: *split-cache
  image: python:3.12
  before_script:
    - curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_linux_amd64.tar.gz | tar xz
    - mv shardfit /usr/local/bin/
  script:
    - shardfit pytest split -n 8 --timings timings.json --out-dir buckets
  artifacts:
    paths: [buckets/]

test:
  stage: test
  parallel:
    matrix: [SHARD: [1, 2, 3, 4, 5, 6, 7, 8]]
  needs: [split]
  script:
    - pytest $(cat buckets/bucket-$SHARD.txt) --junitxml=results-$SHARD.xml
  artifacts:
    paths: [results-*.xml]

report:
  stage: report
  <<: *split-cache
  needs: [test]
  script:
    - shardfit pytest report --junit-xml "results-*.xml" --timings timings.json
```

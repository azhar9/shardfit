# Jenkins

Plain pipeline with a parallel stage per bucket. Store the timing file in
the workspace and archive it so the next build starts warm.

```groovy
pipeline {
  agent any
  stages {
    stage('Split') {
      steps {
        sh '''curl -sSL https://github.com/azhar9/shardfit/releases/latest/download/shardfit_Linux_x86_64.tar.gz | tar xz
              ./shardfit pytest split -n 8 --timings timings.json --out-dir buckets'''
      }
    }
    stage('Test') {
      steps {
        script {
          def shards = [:]
          for (int i = 1; i <= 8; i++) {
            def s = i
            shards["shard-${s}"] = {
              sh "pytest \$(cat buckets/bucket-${s}.txt) --junitxml=results-${s}.xml"
            }
          }
          parallel shards
        }
      }
    }
    stage('Report') {
      steps {
        sh './shardfit pytest report --junit-xml "results-*.xml" --timings timings.json'
      }
    }
  }
  post {
    always {
      archiveArtifacts artifacts: 'timings.json', allowEmptyArchive: true
    }
  }
}
```

Note: the `report` stage runs after `Test` even on failures — put it in
`post { always { ... } }` if your Jenkins version can't express that with
plain stages.

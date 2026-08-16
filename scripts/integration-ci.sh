#!/usr/bin/env bash
# Live adapter integration checks for CI: runs the real pytest and jest
# flows end to end (discover -> split -> run -> report -> informed split).
# Fails loudly on any step; no external fixtures needed.
set -euo pipefail

BIN="${1:-./shardfit}"
# resolve to the physical path: macOS symlinks /var to /private/var and
# jest emits physical paths, which would defeat id relativization
ROOT="$(cd "$(mktemp -d)" && pwd -P)"
cd "$ROOT"

echo "== pytest flow =="
python3 -m venv venv
./venv/bin/pip install -q pytest
export PATH="$PWD/venv/bin:$PATH"
mkdir -p tests
cat > tests/test_a.py <<'EOF'
import time
def test_fast_one(): time.sleep(0.1)
def test_fast_two(): time.sleep(0.2)
EOF
cat > tests/test_b.py <<'EOF'
import time
def test_slow(): time.sleep(1.0)
EOF
N=$("$BIN" pytest discover | wc -l | tr -d ' ')
[ "$N" = "3" ] || { echo "pytest discover: expected 3 tests, got $N"; exit 1; }
"$BIN" pytest split -n 2 --out-dir buckets
[ -s buckets/bucket-1.txt ] && [ -s buckets/bucket-2.txt ] || { echo "pytest split: bucket files missing"; exit 1; }
TOTAL=$(cat buckets/bucket-*.txt | wc -l | tr -d ' ')
[ "$TOTAL" = "3" ] || { echo "pytest split: expected 3 ids, got $TOTAL"; exit 1; }
pytest $(cat buckets/bucket-1.txt) --junitxml=r1.xml -q > /dev/null
pytest $(cat buckets/bucket-2.txt) --junitxml=r2.xml -q > /dev/null
"$BIN" pytest report --junit-xml "r*.xml" --timings timings.json 2>&1 | grep -q "merged 3 durations"
# regression guard: informed split must succeed and report an imbalance
"$BIN" pytest split -n 2 --timings timings.json --estimate-only 2>&1 | grep -q "imbalance:"

echo "== jest flow =="
mkdir -p js && cd js
npm init -y > /dev/null
npm i --no-audit --no-fund jest jest-junit > /dev/null
export PATH="$PWD/node_modules/.bin:$PATH"
mkdir -p src/__tests__
cat > src/__tests__/a.test.js <<'EOF'
test('a fast', () => {});
EOF
cat > src/__tests__/b.test.js <<'EOF'
test('b slow', () => { const end = Date.now() + 1200; while (Date.now() < end); });
EOF
cat > jest.config.js <<'EOF'
module.exports = { reporters: ["default", ["jest-junit", { classNameTemplate: "{filepath}" }]] };
EOF
N=$("$BIN" jest discover | wc -l | tr -d ' ')
[ "$N" = "2" ] || { echo "jest discover: expected 2 files, got $N"; exit 1; }
"$BIN" jest split -n 2 --out-dir buckets
jest $(cat buckets/bucket-1.txt) > /dev/null 2>&1; cp junit.xml j1.xml
jest $(cat buckets/bucket-2.txt) > /dev/null 2>&1; cp junit.xml j2.xml
"$BIN" jest report --junit-xml "j*.xml" --timings timings.json 2>&1 | grep -q "merged 2 durations"
# regression guard: with the correct classNameTemplate the informed split
# succeeds; with title-based classnames it errors on 100% unknown ids
"$BIN" jest split -n 2 --timings timings.json --estimate-only 2>&1 | grep -q "imbalance:"

echo "integration OK"

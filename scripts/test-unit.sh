#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
mkdir -p .artifacts/examples-coverage
find .artifacts/examples-coverage -type f -delete
export VALIDATION_EXAMPLE_COVERAGE_DIR="$repo_root/.artifacts/examples-coverage"
go test -count=1 -json -covermode=atomic -coverpkg=./... -coverprofile=.artifacts/runtime.out ./... | tee .artifacts/runtime-tests.json
go tool covdata textfmt -i=.artifacts/examples-coverage -o=.artifacts/examples.out
go -C docs test -count=1 -json -covermode=atomic -coverprofile="$repo_root/.artifacts/docs.out" ./... | tee .artifacts/docs-tests.json
python3 -m unittest discover -s scripts -p '*_test.py'
python3 scripts/report.py

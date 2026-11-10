#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
snapshot="$(mktemp)"
currency_snapshot="$(mktemp)"
trap 'rm -f "$snapshot" "$currency_snapshot"' EXIT
cp README.md "$snapshot"
cp currency_codes.go "$currency_snapshot"
python3 scripts/currencies.py
go -C docs run ./readme
cmp README.md "$snapshot"
cmp currency_codes.go "$currency_snapshot"
python3 scripts/currencies.py
go -C docs run ./readme
cmp README.md "$snapshot"
cmp currency_codes.go "$currency_snapshot"
python3 - <<'PY'
import json
import sys
from pathlib import Path
sys.path.insert(0, 'scripts')
from report import fingerprint
report = json.loads(Path('docs/quality.json').read_text())
if report['source_sha256'] != fingerprint(Path('.')):
    raise SystemExit('unit evidence is stale; run make test-coverage')
PY

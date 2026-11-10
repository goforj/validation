#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
revision="${1:-$(git -C "$repo_root" rev-parse HEAD)}"
consumer_dir="$(mktemp -d /tmp/validation-consumer.XXXXXX)"
# cleanup uses Go's remover because downloaded module directories are read-only.
cleanup() {
  local consumer_status=$?
  trap - EXIT
  for cache in modcache lookupcache; do
    if ! GOMODCACHE="$consumer_dir/$cache" go clean -modcache; then consumer_status=1; fi
  done
  if ! rm -rf "$consumer_dir"; then consumer_status=1; fi
  exit "$consumer_status"
}
trap cleanup EXIT
cp -R "$repo_root/examples" "$consumer_dir/examples"
cd "$consumer_dir"
export GOENV=off GOWORK=off GOFLAGS=""
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org
export GOPRIVATE="" GONOPROXY="" GONOSUMDB="" GOINSECURE=""
export GOMODCACHE="$consumer_dir/modcache"
unset GOCOVERDIR VALIDATION_EXAMPLE_COVERAGE_DIR
go mod init example.com/validation-consumer
# Resolve metadata separately so a cached proxy lookup failure can fall back to GitHub.
# Installation must still fetch the canonical version through the public proxy alone.
version="$(GOMODCACHE="$consumer_dir/lookupcache" GOPROXY=https://proxy.golang.org,direct go list -m -f '{{.Version}}' "github.com/goforj/validation@$revision")"
test -n "$version"
go get "github.com/goforj/validation@$version"
go list -m -json all > modules.json
python3 - "$version" <<'PY'
import json
import sys
from pathlib import Path

source = Path('modules.json').read_text()
decoder = json.JSONDecoder()
selected = None
while source.strip():
    module, end = decoder.raw_decode(source.lstrip())
    source = source.lstrip()[end:]
    if 'Replace' in module:
        raise SystemExit('consumer unexpectedly resolved a local replacement')
    if module['Path'] == 'github.com/goforj/validation':
        selected = module['Version']
if not selected:
    raise SystemExit('consumer did not resolve the validation module')
if selected != sys.argv[1]:
    raise SystemExit('consumer resolved a different module version')
print('Public module:', selected)
PY
go test -count=1 ./examples/...

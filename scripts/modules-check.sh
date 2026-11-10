#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
actual="$(find . -name go.mod -not -path './.artifacts/*' -not -path './.git/*' | sed 's|^./||' | sort)"
expected="$(sort scripts/module-manifest.txt)"
if [[ "$actual" != "$expected" ]]; then
  printf 'module inventory differs from scripts/module-manifest.txt\n' >&2
  exit 1
fi

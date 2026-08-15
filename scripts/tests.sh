#!/usr/bin/env bash
set -euo pipefail

go_command="go"
if ! command -v go >/dev/null 2>&1; then
  portable_go="../.toolchains/go1.22.12/go/bin"
  if [[ ! -x "$portable_go/go.exe" ]]; then
    echo "Go 1.22.12 toolchain not found"
    exit 1
  fi
  go_command="$portable_go/go.exe"
fi

"$go_command" test ./...
"$go_command" vet ./...
node scripts/build.mjs --clean
node --test --experimental-strip-types "tests/node/*.test.ts"

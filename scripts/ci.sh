#!/usr/bin/env bash
set -euo pipefail

gofmt -w src
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git diff --exit-code -- src
fi
go test ./...
go vet ./...
node scripts/build.mjs
node --test --experimental-strip-types "tests/node/*.test.ts"
node scripts/check-loc.mjs

#!/usr/bin/env bash
set -euo pipefail

go_command="go"
gofmt_command="gofmt"
if ! command -v go >/dev/null 2>&1; then
  portable_go="../.toolchains/go1.22.12/go/bin"
  if [[ ! -x "$portable_go/go.exe" ]]; then
    echo "Go 1.22.12 toolchain not found"
    exit 1
  fi
  go_command="$portable_go/go.exe"
  gofmt_command="$portable_go/gofmt.exe"
fi

format_scope="src"
if [[ "$go_command" == *.exe || "${OSTYPE:-}" == msys* ]]; then
  format_scope="src/solvency"
fi
unformatted="$("$gofmt_command" -l "$format_scope")"
if [[ -n "$unformatted" ]]; then
  echo "Go files require gofmt:"
  echo "$unformatted"
  exit 1
fi

"$go_command" test ./...
"$go_command" vet ./...
node scripts/build.mjs --clean
node --test --experimental-strip-types "tests/node/*.test.ts"
npx prettier --check README.md SECURITY.md docs sdk tests package.json package-lock.json .github
npm audit --audit-level=high
node scripts/check-loc.mjs
node scripts/verify-release.mjs

if [[ -z "${WSL_DISTRO_NAME:-}" ]] && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git diff --exit-code
fi

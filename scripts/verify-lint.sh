#!/usr/bin/env bash
# Enforce zero lint issues before verify completes.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TEST_PKGS=(
	./tests/...
	./internal/...
	./pkg/...
	./cmd/...
)

echo "==> go vet"
go vet "${TEST_PKGS[@]}"

if command -v golangci-lint >/dev/null 2>&1; then
	echo "==> golangci-lint"
	golangci-lint run "${TEST_PKGS[@]}"
else
	echo "==> golangci-lint not installed; go vet is the lint gate"
fi

echo "==> lint gate passed"

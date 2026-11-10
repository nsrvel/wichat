#!/usr/bin/env bash
# Enforces scoped unit-test coverage for a backend module (service + pkg + repository/mock).
# Target: 90% (see docs/architecture/backend-quality.md). Floor: .coverage-min in module root.
set -euo pipefail

MODULE_DIR="${1:?module dir}"
TARGET="${2:-90}"
cd "$MODULE_DIR"

if ! command -v go >/dev/null; then
  echo "coverage-gate: go not found" >&2
  exit 1
fi

COVER_PKGS=()
if [[ -f coverage.scope ]]; then
  SCOPE=$(grep -v '^\s*#' coverage.scope | tr '\n' ' ')
  while IFS= read -r pkg; do
    [[ -n "$pkg" ]] && COVER_PKGS+=("$pkg")
  done < <(go list $SCOPE 2>/dev/null | sort -u)
else
  while IFS= read -r pkg; do
    [[ -n "$pkg" ]] && COVER_PKGS+=("$pkg")
  done < <(go list ./internal/.../service ./pkg/... ./internal/.../repository/mock 2>/dev/null | sort -u)
fi

if ((${#COVER_PKGS[@]} == 0)); then
  while IFS= read -r pkg; do
    [[ -n "$pkg" ]] && COVER_PKGS+=("$pkg")
  done < <(go list ./pkg/... ./internal/middleware ./internal/realtime/hub 2>/dev/null | sort -u)
fi

if ((${#COVER_PKGS[@]} == 0)); then
  echo "coverage-gate: no scoped packages under $MODULE_DIR" >&2
  exit 1
fi

COVERPKG=$(IFS=,; echo "${COVER_PKGS[*]}")
PROFILE="$(mktemp)"
trap 'rm -f "$PROFILE"' EXIT

set +e
go test ./... -coverpkg="$COVERPKG" -coverprofile="$PROFILE" >/dev/null 2>&1
TEST_EXIT=$?
set -e
if [[ "$TEST_EXIT" -ne 0 ]]; then
  echo "coverage-gate: go test failed in $MODULE_DIR" >&2
  go test ./... -coverpkg="$COVERPKG" -coverprofile="$PROFILE"
  exit 1
fi

ACTUAL=$(go tool cover -func="$PROFILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')
FLOOR="$TARGET"
if [[ -f .coverage-min ]]; then
  FLOOR=$(tr -d '[:space:]' < .coverage-min)
fi

awk -v actual="$ACTUAL" -v floor="$FLOOR" 'BEGIN {
  if (actual+0 < floor+0) { exit 1 }
  exit 0
}' || {
  echo "coverage-gate: $(basename "$MODULE_DIR") scoped coverage ${ACTUAL}% < floor ${FLOOR}% (target ${TARGET}%)" >&2
  go tool cover -func="$PROFILE" | tail -1
  exit 1
}

echo "coverage-gate: $(basename "$MODULE_DIR") scoped ${ACTUAL}% (floor ${FLOOR}%, target ${TARGET}%)"

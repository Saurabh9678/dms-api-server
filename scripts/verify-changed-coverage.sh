#!/usr/bin/env bash
# Require 100% statement coverage on production packages touched since the base ref.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

MODULE="$(go list -m)"
BASE_REF="${VERIFY_BASE_REF:-}"
if [[ -z "$BASE_REF" ]]; then
	if git rev-parse --verify origin/main >/dev/null 2>&1; then
		BASE_REF="origin/main"
	elif git rev-parse --verify main >/dev/null 2>&1; then
		BASE_REF="main"
	else
		BASE_REF="HEAD"
	fi
fi

PACKAGE_LIST="$(mktemp)"
trap 'rm -f "$PACKAGE_LIST"' EXIT

add_package_dir() {
	local dir="$1"
	if ! grep -Fxq "$dir" "$PACKAGE_LIST" 2>/dev/null; then
		echo "$dir" >>"$PACKAGE_LIST"
	fi
}

collect_changed_go_files() {
	{
		git diff --name-only --diff-filter=ACMR "${BASE_REF}"...HEAD 2>/dev/null || true
		git diff --name-only --diff-filter=ACMR "${BASE_REF}" HEAD 2>/dev/null || true
		git diff --name-only --diff-filter=ACMR 2>/dev/null || true
		git diff --name-only --diff-filter=ACMR --cached 2>/dev/null || true
	} | sort -u | grep -E '\.go$' || true
}

map_changed_file_to_packages() {
	local file="$1"

	if [[ "$file" =~ ^(internal|pkg|cmd)/.*\.go$ ]] && [[ ! "$file" =~ _test\.go$ ]]; then
		add_package_dir "$(dirname "$file")"
	fi

	if [[ "$file" =~ ^tests/unit/([^/]+)/ ]]; then
		local module_name="${BASH_REMATCH[1]}"
		if [[ -d "internal/modules/${module_name}" ]]; then
			add_package_dir "internal/modules/${module_name}"
		fi
	fi
}

while IFS= read -r file; do
	[[ -z "$file" ]] && continue
	map_changed_file_to_packages "$file"
done < <(collect_changed_go_files)

if [[ "${VERIFY_COVERAGE_ALL:-}" == "1" ]]; then
	for module_dir in internal/modules/*/; do
		[[ -d "$module_dir" ]] || continue
		add_package_dir "${module_dir%/}"
	done
fi

if [[ ! -s "$PACKAGE_LIST" ]]; then
	echo "==> coverage gate: no changed production packages (base: ${BASE_REF})"
	exit 0
fi

resolve_test_targets() {
	local pkg_dir="$1"
	local targets=""

	case "$pkg_dir" in
	internal/modules/*)
		local module_name="${pkg_dir#internal/modules/}"
		if [[ -d "tests/unit/${module_name}" ]]; then
			targets="./tests/unit/${module_name}/..."
		fi
		;;
	esac

	if [[ -z "$targets" ]]; then
		targets="./tests/..."
	fi

	echo "$targets"
}

assert_full_coverage() {
	local pkg_dir="$1"
	local profile="$2"
	local import_prefix="${MODULE}/${pkg_dir}"
	local failed=0

	while IFS= read -r line; do
		[[ -z "$line" ]] && continue
		local pct
		pct="$(awk '{print $NF}' <<<"$line" | tr -d '%')"
		if awk -v value="$pct" 'BEGIN { exit !(value + 0 < 100) }'; then
			echo "  uncovered: ${line}"
			failed=1
		fi
	done < <(go tool cover -func="$profile" | grep "${import_prefix}/" || true)

	if ! go tool cover -func="$profile" | grep -q "${import_prefix}/"; then
		echo "  no coverage data recorded for ${pkg_dir}"
		failed=1
	fi

	if [[ "$failed" -ne 0 ]]; then
		echo "FAIL: ${pkg_dir} requires 100% coverage on every function"
		go tool cover -func="$profile" | grep "${import_prefix}/" || true
		return 1
	fi

	echo "PASS: ${pkg_dir} has 100% function coverage"
	return 0
}

echo "==> coverage gate: base ref ${BASE_REF}"

OVERALL_FAILED=0
PROFILE_DIR="$(mktemp -d)"
trap 'rm -rf "$PROFILE_DIR"; rm -f "$PACKAGE_LIST"' EXIT

while IFS= read -r pkg_dir; do
	[[ -z "$pkg_dir" ]] && continue

	coverpkg="./${pkg_dir}/..."
	test_targets="$(resolve_test_targets "$pkg_dir")"
	profile="${PROFILE_DIR}/$(echo "$pkg_dir" | tr '/' '_').out"

	echo "==> checking ${pkg_dir} (tests: ${test_targets})"

	if ! go test -coverprofile="$profile" -coverpkg="$coverpkg" $test_targets; then
		echo "FAIL: tests failed for ${pkg_dir}"
		OVERALL_FAILED=1
		continue
	fi

	if ! assert_full_coverage "$pkg_dir" "$profile"; then
		OVERALL_FAILED=1
	fi
done < <(sort -u "$PACKAGE_LIST")

if [[ "$OVERALL_FAILED" -ne 0 ]]; then
	echo "FAIL: coverage gate requires 100% function coverage on all changed packages"
	exit 1
fi

echo "==> coverage gate passed"

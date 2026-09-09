#!/usr/bin/env bash
#
# Self-test for scripts/vulncheck.sh: feeds it recorded govulncheck JSON and
# asserts how each bucket is classified and which mode fails on it. Run with
# `make vulncheck-test`; CI runs it before the real scan.
set -uo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
script="$here/vulncheck.sh"
data="$here/testdata"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

failures=0

# check NAME EXPECTED_EXIT FIXTURE MODE [EXPECTED_SUBSTRING...]
check() {
  local name="$1" want="$2" fixture="$3" mode="$4"; shift 4
  local out="$tmp/$name.md" got=0
  "$script" --mode "$mode" --json "$data/$fixture" --report "$out" >"$tmp/$name.log" 2>&1 || got=$?
  if [ "$got" -ne "$want" ]; then
    echo "FAIL $name: exit $got, want $want"
    sed 's/^/    /' "$tmp/$name.log"
    failures=$((failures + 1))
    return
  fi
  local needle
  for needle in "$@"; do
    if ! grep -qF -- "$needle" "$out"; then
      echo "FAIL $name: report is missing: $needle"
      sed 's/^/    /' "$out"
      failures=$((failures + 1))
      return
    fi
  done
  echo "ok   $name"
}

check clean-pr           0 clean.json            pr   "no known vulnerabilities"
check clean-full         0 clean.json            full "no known vulnerabilities"
# A stdlib-only disclosure must not fail an unrelated pull request, but must
# fail the scheduled run so it gets an issue and a toolchain bump.
check toolchain-pr       0 toolchain-only.json   pr   "GO-2026-6218" "toolchain" "\`1.26.6\`" "### Toolchain (2)"
check toolchain-full     1 toolchain-only.json   full "### Toolchain (2)"
# A dependency vulnerability fails in every mode, stdlib findings alongside it
# stay classified as toolchain.
check module-pr          1 module-vuln.json      pr   "GO-2025-1234" "\`v0.36.3\`" "k8s.io/apimachinery" "### Module (1)" "### Toolchain (1)"
check module-full        1 module-vuln.json      full "### Module (1)"

# Deeper call frames must not change the bucket: trace[0] is the vulnerable
# symbol, later frames are our own packages.
if grep -q "kuberescue" "$tmp/toolchain-pr.md"; then
  echo "FAIL trace-attribution: caller module leaked into the report"
  failures=$((failures + 1))
else
  echo "ok   trace-attribution"
fi

if [ "$failures" -ne 0 ]; then
  echo "$failures check(s) failed"
  exit 1
fi
echo "all vulncheck checks passed"

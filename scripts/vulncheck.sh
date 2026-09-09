#!/usr/bin/env bash
#
# Runs govulncheck and splits the findings into two buckets:
#
#   toolchain  vulnerabilities in the Go standard library, fixed by bumping
#              the `go` directive in go.mod to the fixed release
#   module     vulnerabilities in a dependency or in our own code
#
# The split exists because the two need different responses. A stdlib CVE can
# be published at any moment against the pinned toolchain and has nothing to do
# with the pull request that happens to run next, so on pull requests it is
# reported and not fatal (--mode pr). The scheduled run is strict (--mode full)
# and files an issue, which is where toolchain bumps get tracked.
#
# Usage:
#   scripts/vulncheck.sh [--mode pr|full] [--json FILE] [--report FILE]
#
#   --mode pr      fail only on module vulnerabilities (default)
#   --mode full    fail on any vulnerability
#   --json FILE    read govulncheck JSON from FILE instead of scanning
#   --report FILE  write the markdown report here as well as to stdout
#
# Exit codes: 0 nothing to fail on, 1 vulnerabilities that fail this mode,
# 2 usage or tooling error.
set -euo pipefail

mode="pr"
json_in=""
report_out=""

die() { echo "vulncheck: $*" >&2; exit 2; }

while [ $# -gt 0 ]; do
  case "$1" in
    --mode)   mode="${2:-}"; shift 2 || die "--mode needs a value" ;;
    --json)   json_in="${2:-}"; shift 2 || die "--json needs a value" ;;
    --report) report_out="${2:-}"; shift 2 || die "--report needs a value" ;;
    -h|--help) sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) die "unknown argument: $1" ;;
  esac
done

case "$mode" in
  pr|full) ;;
  *) die "--mode must be pr or full, got: $mode" ;;
esac

command -v jq >/dev/null 2>&1 || die "jq is required (brew install jq / apt-get install jq)"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

json="$workdir/govulncheck.json"
if [ -n "$json_in" ]; then
  [ -f "$json_in" ] || die "no such file: $json_in"
  cp "$json_in" "$json"
else
  command -v go >/dev/null 2>&1 || die "go is required"
  # Scan with the toolchain go.mod pins, not whatever the newest installed
  # release is: a newer stdlib hides exactly the CVEs this is meant to catch.
  # "+auto" still allows an upgrade if something genuinely requires one.
  pin="$(awk '$1 == "go" { print $2; exit }' go.mod 2>/dev/null || true)"
  if [ -n "$pin" ]; then
    export GOTOOLCHAIN="go${pin}+auto"
  fi
  # -format json exits 0 even when it finds something, so a non-zero exit here
  # is a tooling failure (bad packages, no network) and not a vulnerability.
  if command -v govulncheck >/dev/null 2>&1; then
    govulncheck -format json ./... >"$json" || die "govulncheck failed to run"
  else
    go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION:-latest}" \
      -format json ./... >"$json" || die "govulncheck failed to run"
  fi
fi

# One row per vulnerability: id, bucket, fixed version, affected packages.
# trace[0] is the vulnerable symbol's own frame, so its module identifies the
# vulnerable module ("stdlib" for the standard library and the toolchain).
jq -s '
  [ .[] | select(has("finding")) | .finding ]
  | map(select(.trace != null and (.trace | length) > 0))
  | group_by(.osv)
  | map({
      id: .[0].osv,
      bucket: (if (.[0].trace[0].module // "") == "stdlib" then "toolchain" else "module" end),
      module: (.[0].trace[0].module // "unknown"),
      fixed: (map(.fixed_version // empty) | first // "none"),
      packages: (([ .[] | .trace[0].package // empty ] | unique | join(", ")) as $p
                 | if $p == "" then "-" else $p end)
    })
  # govulncheck reports stdlib fixes as "v1.26.6"; go.mod wants "1.26.6".
  | map(if .bucket == "toolchain"
        then .fixed |= (sub("^v"; "") | sub("^go"; ""))
        else . end)
  | sort_by(.bucket, .id)
' "$json" >"$workdir/vulns.json" || die "could not parse govulncheck output"

count_of() { jq --arg b "$1" '[ .[] | select(.bucket == $b) ] | length' "$workdir/vulns.json"; }
toolchain_count="$(count_of toolchain)"
module_count="$(count_of module)"

go_version="$(jq -rs '[ .[] | select(has("config")) | .config.go_version ] | first // "unknown"' "$json")"

# Only meaningful for a scan we ran ourselves; a recorded --json file carries
# whatever toolchain it was recorded with.
toolchain_note=""
if [ -z "$json_in" ] && [ -n "${pin:-}" ] && [ "$go_version" != "go$pin" ]; then
  toolchain_note="Scanned with \`$go_version\`, but go.mod pins \`$pin\` — these results may not match what CI builds."
fi

report="$workdir/report.md"
{
  if [ "$toolchain_count" -eq 0 ] && [ "$module_count" -eq 0 ]; then
    echo "## govulncheck: no known vulnerabilities"
    echo
    echo "Scanned with the toolchain in \`go.mod\` (\`$go_version\`)."
    if [ -n "$toolchain_note" ]; then echo; echo "> [!WARNING]"; echo "> $toolchain_note"; fi
  else
    echo "## govulncheck findings"
    echo
    echo "Scanned with the toolchain in \`go.mod\` (\`$go_version\`)."
    if [ -n "$toolchain_note" ]; then echo; echo "> [!WARNING]"; echo "> $toolchain_note"; fi
    echo
    echo "| ID | Kind | Vulnerable module | Fixed in | Packages |"
    echo "| --- | --- | --- | --- | --- |"
    jq -r '.[] | "| [\(.id)](https://pkg.go.dev/vuln/\(.id)) | \(.bucket) | `\(.module)` | `\(.fixed)` | \(.packages) |"' "$workdir/vulns.json"
    echo
    if [ "$toolchain_count" -gt 0 ]; then
      echo "### Toolchain ($toolchain_count)"
      echo
      echo "Standard library vulnerabilities. Fix by setting the \`go\` directive"
      echo "in \`go.mod\` to the \`Fixed in\` release above (highest one wins), then"
      echo "re-running this script. No source change is needed."
    fi
    if [ "$module_count" -gt 0 ]; then
      echo
      echo "### Module ($module_count)"
      echo
      echo "Vulnerabilities in a dependency or in our own code. Fix by upgrading"
      echo "the module (\`go get -u <module>\` plus \`go mod tidy\`) or by removing"
      echo "the vulnerable call path."
    fi
  fi
} >"$report"

cat "$report"
if [ -n "$report_out" ]; then cp "$report" "$report_out"; fi
if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then cat "$report" >>"$GITHUB_STEP_SUMMARY"; fi
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "toolchain_count=$toolchain_count"
    echo "module_count=$module_count"
    echo "ids=$(jq -r '[ .[] | .id ] | join(" ")' "$workdir/vulns.json")"
  } >>"$GITHUB_OUTPUT"
fi

if [ "$module_count" -gt 0 ]; then
  echo "vulncheck: $module_count module vulnerability(ies) — failing" >&2
  exit 1
fi
if [ "$toolchain_count" -gt 0 ]; then
  if [ "$mode" = full ]; then
    echo "vulncheck: $toolchain_count toolchain vulnerability(ies) — failing (mode=full)" >&2
    exit 1
  fi
  echo "vulncheck: $toolchain_count toolchain vulnerability(ies) — not failing this pull request (mode=pr)" >&2
fi
exit 0

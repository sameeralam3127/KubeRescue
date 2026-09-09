# Development

## Prerequisites

- Go 1.24+ (the toolchain declared in `go.mod` is downloaded automatically)
- Docker (optional, for image builds)
- A cluster to test against — [kind](https://kind.sigs.k8s.io/) works well

## Build and test

```bash
make build      # produces bin/kuberescue
make test       # go test -race -cover ./...
make lint       # gofmt check + go vet
```

## Running locally

KubeRescue uses your kubeconfig automatically outside a cluster:

```bash
bin/kuberescue monitor -n default --once --dry-run
```

## Project conventions

- Detectors are pure functions over pod state: no API calls, no mutation.
  They live in `internal/detect` and must be testable without a cluster.
- Remediation actions return a truthful `Outcome` — a dry run is never
  reported as a remediation, and every skip carries its reason.
- Reports go to stdout; logs (structured, via `log/slog`) go to stderr.
- Exit codes: 0 = clean, 1 = error, 2 = findings (in `--once` mode, and in
  `diagnose`).
- All Kubernetes access goes through `kubernetes.Interface` so tests can
  use `k8s.io/client-go/kubernetes/fake`.
- `kuberescue diagnose` (`internal/diagnose`) is read-only by design: it
  never calls `internal/remediate`. Explanations are fixed heuristics keyed
  off each finding's `Reason`, not a model call — diagnosis stays
  reproducible, and AI stays out of the action path entirely.

## Vulnerability scanning

`scripts/vulncheck.sh` runs govulncheck with JSON output and buckets each
finding by the module of `trace[0]` — the vulnerable symbol's own frame.
`stdlib` findings are **toolchain** vulnerabilities, everything else is a
**module** vulnerability.

The split is there because the two fail differently. A toolchain CVE is
published against whatever version `go.mod` pins, on a schedule nobody here
controls, so making it fail pull requests means unrelated changes get
blocked by a one-line `go.mod` bump (this happened twice: commit 41b9386 and
issue #24). So `--mode pr`, used by CI, fails only on module findings and
reports toolchain ones in the job summary; `--mode full`, used by the
scheduled `Vulncheck` workflow and by `make vulncheck`, fails on both and
files an issue labelled `vulncheck` that closes itself when a later run is
clean.

`make vulncheck-test` runs the classifier against recorded govulncheck JSON
in `scripts/testdata/` — including a stdlib finding whose call stack reaches
into our own packages, which must still bucket as toolchain.

## Releasing

Push a `vX.Y.Z` tag and `.github/workflows/release.yml` takes it from
there: goreleaser cross-builds the binaries, publishes the GitHub release
with checksums, and pushes the multi-arch `ghcr.io` image. Validate the
config locally first with `goreleaser check` and
`goreleaser release --snapshot --clean --skip=publish,docker`.

Homebrew (`homebrew_casks`) and krew (`krews`) publishing are written into
`.goreleaser.yml` as a comment but not enabled: both push to a separate
repo (`sameeralam3127/homebrew-kuberescue`, `sameeralam3127/krew-index` —
both created, currently empty) using a `HOMEBREW_TAP_GITHUB_TOKEN` secret
that has to be a fine-grained PAT minted by hand at
github.com/settings/personal-access-tokens/new (GitHub doesn't allow
minting PATs via API). Once that secret exists on this repo, add a `homebrew_casks` block (name,
repository owner/name/token, homepage, description, `binaries: [kuberescue]`)
and a `krews` block (same shape, plus `short_description`) to
`.goreleaser.yml` in place of that comment, and the next tag will start
publishing both.

## Docker image

```bash
make docker
docker run --rm kuberescue:local --help
```

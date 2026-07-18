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
- Exit codes: 0 = clean, 1 = error, 2 = findings (in `--once` mode).
- All Kubernetes access goes through `kubernetes.Interface` so tests can
  use `k8s.io/client-go/kubernetes/fake`.

## Docker image

```bash
make docker
docker run --rm kuberescue:local --help
```

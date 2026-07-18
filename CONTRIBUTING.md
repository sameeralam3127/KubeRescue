# Contributing

Thanks for helping improve KubeRescue.

## Local checks

Before opening a pull request, run:

```bash
make lint   # gofmt + go vet
make test   # go test -race -cover ./...
```

CI additionally runs golangci-lint and govulncheck.

## Pull request guidelines

- Keep changes focused and easy to review.
- Add or update tests for behavior changes; tests use the fake clientset
  from `k8s.io/client-go/kubernetes/fake` — no cluster required.
- Update the README or docs when commands, setup, or behavior changes.
- Avoid adding new dependencies unless they clearly reduce project risk or
  complexity.
- Safety invariants are non-negotiable: actions report truthful outcomes,
  dry-run never mutates, bare pods are never deleted, and the monitor loop
  never dies on a transient API error. See
  [docs/development.md](docs/development.md).

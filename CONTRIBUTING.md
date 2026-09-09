# Contributing

Thanks for helping improve KubeRescue.

## Local checks

Before opening a pull request, run:

```bash
make lint   # gofmt + go vet
make test   # go test -race -cover ./...
```

CI additionally runs golangci-lint and govulncheck.

## When govulncheck fails

`make vulncheck` splits findings into two kinds, and they have different
fixes:

- **toolchain** — a standard library CVE. Set the `go` directive in `go.mod`
  to the highest `Fixed in` version the report lists, run `make vulncheck`
  again, and commit that one-line change. No source change is needed.
- **module** — a vulnerability in a dependency or in our own code. Upgrade
  the module (`go get -u <module> && go mod tidy`) or remove the vulnerable
  call path.

Pull requests only fail on **module** findings. A stdlib CVE can be
disclosed against the pinned toolchain on any day and is unrelated to the
pull request that happens to run next, so those are reported in the job
summary instead. The scheduled `Vulncheck` workflow is strict about both and
opens (and later closes) an issue labelled `vulncheck` to track them.

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

# KubeRescue

> A Kubernetes remediation engine that understands failures before it fixes them.

KubeRescue watches Kubernetes workloads for failure states, collects the
evidence needed to explain them, and takes **safe, bounded, truthful**
remediation actions. Today it handles CrashLoopBackOff; the architecture is
built to grow into a full evidence → diagnosis → policy → action → verify
pipeline (see [Roadmap](#roadmap)).

> ⚠️ **Pre-1.0.** Use in development and staging clusters. Production
> safety features (cooldowns, policy engine, audit trail) land in upcoming
> milestones — until then, run with `--dry-run` anywhere you care about.

---

## Design principles

1. **Safety over automation** — pods without a controller are never
   deleted; every action is budgeted; dry-run is first-class.
2. **Explainability over magic** — every action carries its evidence
   (restart count, exit code, termination reason, owner) and every skip
   carries its reason.
3. **Truthful reporting** — a dry run is never counted as a remediation;
   counters in reports reflect what actually happened.
4. **Resilience** — a transient API error degrades one scan, never the
   process; retries use capped exponential backoff.

## Features

- 🔍 Detects pods stuck in `CrashLoopBackOff` with full evidence
  (restarts, last exit code, termination reason, owning controller)
- 🛡️ Refuses to delete bare pods — deletion only restarts
  controller-managed workloads
- 🧪 `--dry-run` previews every action without touching the cluster
- 📉 `--max-restarts` caps remediations per scan
- 📦 Versioned JSON reports (`schemaVersion: v1alpha1`) for automation;
  reports on stdout, structured logs on stderr
- 🚦 CI-friendly exit codes: `0` clean, `1` error, `2` findings
- ☸️ In-cluster config or local kubeconfig (`--kubeconfig`, `--context`)

## Install

### From source

```bash
git clone https://github.com/sameeralam3127/KubeRescue.git
cd KubeRescue
make build
bin/kuberescue --help
```

### Docker

```bash
make docker
docker run --rm kuberescue:local --help
```

## Quick start

Preview what KubeRescue would do (changes nothing):

```bash
kuberescue monitor -n default --once --dry-run
```

Scan once and remediate, at most 3 restarts:

```bash
kuberescue monitor -n default --once --max-restarts 3
```

Monitor continuously:

```bash
kuberescue monitor -n default --interval 30s --max-restarts 3
```

JSON for automation (exit code 2 signals findings):

```bash
kuberescue monitor -n default --once --dry-run -o json
```

## Example output

```text
CrashLoopBackOff  default/api-7c8f9f6d9b-x2q4m
  container=api restarts=7 lastReason=OOMKilled exitCode=137 owner=ReplicaSet/api-7c8f9f6d9b
  action: restarted

Summary: detected=1 restarted=1 skipped=0 failed=0
```

```json
{
  "schemaVersion": "v1alpha1",
  "timestamp": "2026-07-18T12:00:00Z",
  "namespace": "default",
  "dryRun": true,
  "detected": 1,
  "restarted": 0,
  "skipped": 0,
  "failed": 0,
  "findings": [
    {
      "namespace": "default",
      "pod": "api-7c8f9f6d9b-x2q4m",
      "container": "api",
      "reason": "CrashLoopBackOff",
      "restartCount": 7,
      "lastTerminationReason": "OOMKilled",
      "lastExitCode": 137,
      "ownerKind": "ReplicaSet",
      "ownerName": "api-7c8f9f6d9b"
    }
  ],
  "actions": [{ "pod": "api-7c8f9f6d9b-x2q4m", "outcome": "dry-run" }]
}
```

## CLI

```text
kuberescue monitor [flags]

  -n, --namespace string    namespace to monitor (default "default")
  -l, --selector string     label selector, for example app=api
  -i, --interval duration   time between scans, for example 30s or 2m (default 30s)
      --once                scan once and exit (exit code 2 when findings exist)
      --dry-run             report what would be done without changing the cluster
      --max-restarts int    maximum pods to restart per scan (0 = unlimited)
  -o, --output string       text or json (default "text")

Global: --kubeconfig, --context, --log-level
```

## Try the demo

```bash
kubectl create namespace kuberescue-test
kubectl apply -n kuberescue-test -f examples/crashloop-demo.yaml

kuberescue monitor -n kuberescue-test --once --dry-run   # observe
kuberescue monitor -n kuberescue-test --once             # remediate

kubectl delete namespace kuberescue-test
```

## Run in-cluster

Manifests under [deploy/kubernetes/](deploy/kubernetes/) ship with
**namespaced RBAC** (a Role scoped to the monitored namespace — no
ClusterRole), a non-root distroless image, and `--dry-run` enabled by
default. Review and remove `--dry-run` deliberately.

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/rbac.yaml
kubectl apply -f deploy/kubernetes/deployment.yaml
```

## Architecture

```text
Engine.Scan
  └── Detector.Detect(pod) ──► Finding (evidence)
        └── remediate.RestartPod ──► Result (restarted | dry-run | skipped | failed)
              └── Report (versioned, truthful counters)
```

Detectors are pure functions over pod state — no API calls, trivially
testable. The `Finding` evidence type is the contract shared by detection,
remediation, and reporting; future stages (diagnosis, policy, verification)
slot into the middle without breaking it. See
[docs/project-structure.md](docs/project-structure.md).

## Roadmap

KubeRescue is being built milestone by milestone toward an intelligent,
policy-gated remediation platform:

- **M1 — Diagnose:** `kuberescue diagnose` / `explain` for the five most
  common failure classes (CrashLoopBackOff, OOMKilled, ImagePullBackOff,
  Pending/FailedScheduling, stuck rollouts), with events + log evidence
- **M2 — Safe remediation kernel:** per-cause actions (e.g. rollout undo,
  report-only), policy gate (cooldowns, rate budgets, protected
  namespaces/labels), `simulate` via server-side dry-run, audit history
- **M3 — Operator:** CRDs (`RemediationPolicy`, `RemediationAction`),
  informer-based controller, leader election, Prometheus metrics,
  Kubernetes Events, approval workflow, Helm chart
- **M4 — Ecosystem:** webhook/Slack notifications, Grafana dashboards,
  OpenTelemetry, signed releases + SBOM, optional AI-assisted explanations
  (never in the action path)

## Development

```bash
make build   # bin/kuberescue
make test    # race detector + coverage
make lint    # gofmt + go vet
```

See [docs/development.md](docs/development.md) and
[CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).

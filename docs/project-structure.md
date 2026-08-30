# Project Structure

```text
.
├── cmd/kuberescue/        CLI entry point (main only; logic lives in internal/)
├── internal/
│   ├── cli/               Cobra commands, flags, exit codes
│   ├── detect/            Detector interface + detectors (pure, no API calls)
│   ├── remediate/         Remediation actions with truthful outcomes
│   ├── engine/            Scan/monitor loops, error resilience, Report type
│   ├── diagnose/          Read-only evidence collection + explanation
│   ├── report/            Text and JSON report rendering
│   └── kube/              Kubernetes client construction
├── deploy/kubernetes/     In-cluster manifests (namespaced RBAC by default)
├── docs/                  Contributor and design documentation
├── examples/              Demo workloads for manual testing
├── Dockerfile             Multi-stage build, distroless non-root runtime
├── Makefile               build / test / lint / docker targets
└── go.mod
```

## Data flow

```text
Engine.Scan (kuberescue monitor)
  └── list pods (namespace + selector)
        └── Detector.Detect(pod) -> []Finding      (evidence, read-only)
              └── remediate.RestartPod -> Result   (restarted | dry-run | skipped | failed)
                    └── Report                     (versioned JSON / text, truthful counters)

diagnose.Run (kuberescue diagnose)
  └── list pods + Deployments (namespace + selector)
        ├── Detector.Detect(pod) -> []Finding                    (same detectors as monitor)
        └── RolloutDetector.Detect(deployment) -> []RolloutFinding
              └── EvidenceBundle{Finding, Events, Explanation}   (never calls remediate)
                    └── Report                                   (versioned JSON / text)
```

The `Finding` type in `internal/detect` is the contract detection and
remediation share; `diagnose` reuses it and adds a separate `RolloutFinding`
type for Deployment-level findings, since a Deployment has no single
container to attribute a finding to. Future milestones (policy,
verification) slot between detection and remediation without changing
either shape.

New directories should be added only when they support a real workflow such
as packaging, deployment, examples, or docs.

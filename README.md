# KubeRescue

KubeRescue is a lightweight Kubernetes auto-remediation tool for development
and staging clusters. It watches pods for `CrashLoopBackOff`, reports useful
diagnostic context, and can delete failed pods so Kubernetes recreates them
through their controller.

> KubeRescue is not production-ready yet. Use it only in local, development, or
> staging clusters until cooldowns, retry budgets, policy controls, and audit
> logging are implemented.

## Why Use It

- Find pods stuck in `CrashLoopBackOff`
- See the failing pod, container, restart count, last termination reason, and
  last exit code
- Preview actions with `--dry-run` before deleting anything
- Target workloads by namespace and label selector
- Run once for scripts or continuously as a small remediation loop
- Emit JSON for automation and CI checks

## Quick Start

Create a virtual environment and install the project:

```bash
python -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

Check the CLI:

```bash
kuberescue --help
```

Safely preview CrashLoopBackOff pods in the default namespace:

```bash
kuberescue --namespace default --once --dry-run
```

Run continuous remediation with a per-scan restart limit:

```bash
kuberescue --namespace default --interval 10 --max-restarts 3
```

KubeRescue first tries in-cluster Kubernetes configuration. If that is not
available, it falls back to your local kubeconfig.

## Common Commands

Scan a namespace once:

```bash
kuberescue --namespace default --once
```

Preview without deleting pods:

```bash
kuberescue --namespace default --once --dry-run
```

Target a workload by label:

```bash
kuberescue --namespace default --selector app=api --once --dry-run
```

Run continuously every 5 seconds:

```bash
kuberescue --namespace default --interval 5 --max-restarts 1
```

Generate JSON for automation:

```bash
kuberescue --namespace default --once --dry-run --output json
```

Example JSON shape:

```json
{
  "detected": 1,
  "findings": [
    {
      "container": "api",
      "last_exit_code": 137,
      "last_terminated_reason": "OOMKilled",
      "namespace": "default",
      "pod": "api-7c8f9f6d9b-x2q4m",
      "restart_count": 7,
      "waiting_reason": "CrashLoopBackOff"
    }
  ],
  "namespace": "default",
  "remediated": 1,
  "selector": null
}
```

## CLI Options

| Option | Description |
| --- | --- |
| `--namespace`, `-n` | Kubernetes namespace to scan. Defaults to `default`. |
| `--selector`, `-l` | Kubernetes label selector, such as `app=api`. |
| `--dry-run` | Report matching pods without deleting them. |
| `--once` | Run one scan and exit. Useful for scripts and scheduled jobs. |
| `--interval`, `-i` | Seconds between scans during continuous monitoring. |
| `--max-restarts` | Maximum pods to delete per scan. |
| `--output`, `-o` | Output format: `text` or `json`. |

## What It Does

KubeRescue scans pod container statuses for the waiting reason
`CrashLoopBackOff`. For each matching pod, it prints diagnostic details and,
unless `--dry-run` is set, deletes the pod. Kubernetes then recreates the pod
through the owning Deployment, ReplicaSet, StatefulSet, or other controller.

The diagnostic output is meant to help decide whether restarting is useful or
whether the workload needs a code, configuration, resource, image, or probe fix.

## What It Does Not Do Yet

- Cooldown windows
- Retry budgets across multiple scans
- Policy-based remediation rules
- Slack, webhook, or metrics output
- Helm chart packaging
- Production safety controls

## Docker Image

Build the local image:

```bash
docker build -t kuberescue:local .
```

The Docker build runs the test suite before producing the final runtime image.

Check the CLI:

```bash
docker run --rm kuberescue:local --help
```

## Docker Desktop Kubernetes Demo

This demo creates a safe, isolated namespace and a workload that intentionally
enters `CrashLoopBackOff`.

Prerequisites:

- Docker Desktop is running
- Kubernetes is enabled in Docker Desktop
- `kubectl config current-context` returns `docker-desktop`
- The image `kuberescue:local` exists from the Docker build above

Create the test namespace:

```bash
kubectl create namespace kuberescue-test
```

Create the crashing demo workload:

```bash
kubectl apply -n kuberescue-test -f examples/crashloop-demo.yaml
```

Confirm the pod is in `CrashLoopBackOff`:

```bash
kubectl get pod -n kuberescue-test \
  -o jsonpath='{range .items[*]}{.metadata.name}{" reason="}{.status.containerStatuses[0].state.waiting.reason}{" restarts="}{.status.containerStatuses[0].restartCount}{"\n"}{end}'
```

Preview the remediation:

```bash
kuberescue --namespace kuberescue-test --once --dry-run
```

Run KubeRescue with a restart limit:

```bash
kuberescue --namespace kuberescue-test --interval 5 --max-restarts 1
```

In another terminal, watch the pod get replaced:

```bash
kubectl get pods -n kuberescue-test
```

Clean up:

```bash
kubectl delete namespace kuberescue-test
```

## Kubernetes Manifests

Example manifests live in `deploy/kubernetes`.

Apply the basic resources:

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/rbac.yaml
kubectl apply -f deploy/kubernetes/deployment.yaml
```

The deployment manifest points at `kuberescue:local` and scans the `default`
namespace. Update the image, namespace, label selector, interval, and restart
limit before using it outside local testing.

## Development

Run the unit tests:

```bash
pytest -v
```

Run the full local check set:

```bash
black --check .
ruff check .
flake8 .
mypy kuberescue
bandit -r kuberescue
pytest -v
```

More details are in `docs/development.md`.

## Project Structure

```text
.
├── deploy/kubernetes/     Kubernetes manifests
├── docs/                  Maintainer documentation
├── examples/              Demo workloads
├── kuberescue/             Python package source
├── tests/                 Unit tests
├── Dockerfile             Container build
├── pyproject.toml         Package and tool configuration
└── README.md              User-facing quick start
```

## Roadmap

1. Add cooldown windows and retry budgets
2. Replace polling with Kubernetes watch streams
3. Add policy-based remediation rules
4. Add notifications and metrics
5. Package the project for safer cluster deployment

## Contributing

Contributions are welcome. Please read `CONTRIBUTING.md` before opening a pull
request.

## License

KubeRescue is licensed under the MIT License. See `LICENSE` for details.

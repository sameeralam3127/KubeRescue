# KubeRescue

KubeRescue is a small Kubernetes auto-remediation engine for development and
staging clusters.

The current MVP watches one namespace, detects pods stuck in
`CrashLoopBackOff`, and deletes the failed pod so its controller can create a
replacement.

> KubeRescue is not production-ready yet. Use it only in local, development, or
> staging clusters until retry limits, cooldowns, policies, and audit logging
> are implemented.

## What It Does

- Watches pods in a Kubernetes namespace
- Detects container states with reason `CrashLoopBackOff`
- Deletes the failed pod
- Lets Kubernetes recreate the pod through its Deployment, ReplicaSet, or other controller

## What It Does Not Do Yet

- Retry budgets
- Cooldown windows
- Policy-based remediation
- Slack, webhook, or metrics output
- Helm chart packaging
- Production safety controls

## Quick Start

Create a virtual environment and install the project:

```bash
python -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

Run the unit tests:

```bash
pytest -v
```

Preview what KubeRescue would remediate in the default namespace:

```bash
kuberescue --namespace default --once --dry-run
```

Run KubeRescue continuously:

```bash
kuberescue --namespace default --interval 10 --max-restarts 3
```

KubeRescue first tries in-cluster Kubernetes configuration. If that is not
available, it falls back to your local kubeconfig.

Each scan reports the crashing pod, container name, restart count, last
termination reason, and last exit code. Those fields make it easier to decide
whether the pod should be restarted or whether the workload needs a code,
configuration, resource, or probe fix.

Useful CLI options:

- `--namespace`, `-n`: namespace to scan
- `--dry-run`: print matching pods without deleting them
- `--once`: scan once and exit, useful for checks and scheduled jobs
- `--interval`, `-i`: seconds between scans during continuous monitoring
- `--max-restarts`: maximum pods to delete per scan
- `--selector`, `-l`: scan only pods that match a Kubernetes label selector
- `--output`, `-o`: print human-readable text or machine-readable JSON

Target a workload by label:

```bash
kuberescue --namespace default --selector app=api --once --dry-run
```

Generate JSON for automation:

```bash
kuberescue --namespace default --once --dry-run --output json
```

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

The deployment manifest is intentionally simple and points at `kuberescue:local`.
Update the image and namespace arguments before using it outside local testing.

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
└── README.md              Quick start and user guide
```

## Development

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

## Roadmap

1. Add retry limits and cooldown windows
2. Replace polling with Kubernetes watch streams
3. Add policy-based remediation rules
4. Add notifications and metrics
5. Package the project for safer cluster deployment

## Contributing

Contributions are welcome. Please read `CONTRIBUTING.md` before opening a pull
request.

## License

KubeRescue is licensed under the MIT License. See `LICENSE` for details.

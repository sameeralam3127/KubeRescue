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

Run KubeRescue against the default namespace:

```bash
kuberescue --namespace default
```

KubeRescue first tries in-cluster Kubernetes configuration. If that is not
available, it falls back to your local kubeconfig.

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

Run KubeRescue:

```bash
kuberescue --namespace kuberescue-test
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

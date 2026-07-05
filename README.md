# KubeRescue

> Lightweight Kubernetes CrashLoopBackOff auto-remediation for development and staging clusters.

KubeRescue continuously watches Kubernetes pods for **CrashLoopBackOff**, provides
useful diagnostic information, and can automatically delete unhealthy pods so
their controller recreates them.

> ⚠️ **Experimental**
>
> KubeRescue is intended **only for development and staging environments**.
> It is **not production-ready** until safety features such as cooldowns,
> retry budgets, policy controls, and audit logging are implemented.

---

## Features

- 🔍 Detect pods stuck in `CrashLoopBackOff`
- 📊 Display restart count, exit code, and termination reason
- 🧪 Preview actions using `--dry-run`
- 🎯 Filter workloads by namespace and label selector
- 🔄 Run once or continuously
- 📦 JSON output for CI/CD and automation
- ☸️ Works with in-cluster config or local kubeconfig

---

## How It Works

```text
                +------------------------+
                | Kubernetes Cluster     |
                +-----------+------------+
                            |
                            v
                  Watch Pod Status
                            |
                            v
                 CrashLoopBackOff?
                     /        \
                   No          Yes
                   |            |
                   |      Collect diagnostics
                   |            |
                   |      Print information
                   |            |
                   |      --dry-run ?
                   |         /      \
                   |      Yes        No
                   |       |          |
                   |   Report only    |
                   |                  |
                   +-------------> Delete Pod
                                      |
                                      v
                          Controller recreates Pod
```

---

# Installation

## From Source

```bash
git clone https://github.com/<your-org>/kuberescue.git
cd kuberescue

python -m venv .venv
source .venv/bin/activate

pip install -e ".[dev]"
```

Verify installation:

```bash
kuberescue --help
```

---

# Quick Start

Preview CrashLoopBackOff pods:

```bash
kuberescue \
  --namespace default \
  --once \
  --dry-run
```

Automatically remediate:

```bash
kuberescue \
  --namespace default \
  --interval 10 \
  --max-restarts 3
```

KubeRescue first attempts **in-cluster Kubernetes configuration** and falls back
to your local kubeconfig automatically.

---

# Example Output

```text
Namespace : default
Pod       : api-7c8f9f6d9b-x2q4m
Container : api

Status
------
Waiting Reason      CrashLoopBackOff
Restart Count       7
Exit Code           137
Last Termination    OOMKilled

Action
------
Deleted pod (controller will recreate it)
```

JSON output:

```json
{
  "detected": 1,
  "remediated": 1,
  "namespace": "default",
  "selector": null,
  "findings": [
    {
      "pod": "api-7c8f9f6d9b-x2q4m",
      "container": "api",
      "waiting_reason": "CrashLoopBackOff",
      "restart_count": 7,
      "last_terminated_reason": "OOMKilled",
      "last_exit_code": 137
    }
  ]
}
```

---

# Common Commands

Scan once

```bash
kuberescue -n default --once
```

Preview only

```bash
kuberescue -n default --once --dry-run
```

Target a Deployment

```bash
kuberescue \
  -n default \
  -l app=api \
  --once \
  --dry-run
```

Continuous monitoring

```bash
kuberescue \
  -n default \
  --interval 5 \
  --max-restarts 1
```

Automation-friendly JSON

```bash
kuberescue \
  -n default \
  --once \
  --output json
```

---

# CLI Options

| Option            | Description                      |
| ----------------- | -------------------------------- |
| `-n, --namespace` | Namespace to scan                |
| `-l, --selector`  | Label selector                   |
| `--dry-run`       | Preview only                     |
| `--once`          | Scan once and exit               |
| `-i, --interval`  | Scan interval in seconds         |
| `--max-restarts`  | Maximum pods remediated per scan |
| `-o, --output`    | `text` or `json`                 |

---

# Docker

Build

```bash
docker build -t kuberescue:local .
```

Run

```bash
docker run --rm kuberescue:local --help
```

The Docker build automatically executes the unit tests before producing the
runtime image.

---

# Demo

A complete CrashLoopBackOff demonstration is included.

Create demo namespace:

```bash
kubectl create namespace kuberescue-test
```

Deploy the crashing workload:

```bash
kubectl apply \
  -n kuberescue-test \
  -f examples/crashloop-demo.yaml
```

Preview remediation:

```bash
kuberescue \
  -n kuberescue-test \
  --once \
  --dry-run
```

Run continuously:

```bash
kuberescue \
  -n kuberescue-test \
  --interval 5 \
  --max-restarts 1
```

Cleanup:

```bash
kubectl delete namespace kuberescue-test
```

---

# Kubernetes Deployment

Example manifests are available under:

```
deploy/kubernetes/
```

Apply them:

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/rbac.yaml
kubectl apply -f deploy/kubernetes/deployment.yaml
```

Update the Deployment manifest before use:

- image
- namespace
- label selector
- interval
- restart limit

---

# Project Structure

```text
.
├── deploy/
│   └── kubernetes/
├── docs/
├── examples/
├── kuberescue/
├── tests/
├── Dockerfile
├── pyproject.toml
└── README.md
```

---

# Development

Run tests:

```bash
pytest -v
```

Run all quality checks:

```bash
black --check .
ruff check .
flake8 .
mypy kuberescue
bandit -r kuberescue
pytest -v
```

See `docs/development.md` for contributor documentation.

---

# Roadmap

- [ ] Cooldown windows
- [ ] Retry budgets
- [ ] Kubernetes Watch API
- [ ] Policy-based remediation
- [ ] Slack/Webhook notifications
- [ ] Prometheus metrics
- [ ] Helm Chart
- [ ] Production safety controls

---

# Contributing

Contributions are welcome!

Please read **CONTRIBUTING.md** before opening a pull request.

---

# License

Licensed under the **MIT License**.

See **LICENSE** for details.

### KubeRescue

> ⚠️ **Status: Alpha – Under Active Development. Not Production Ready.**

Autonomous Kubernetes failure detection and policy-driven auto-remediation engine for SRE and platform engineering teams.

KubeRescue is designed to reduce operational toil by automatically detecting workload failures and applying safe, policy-driven remediation strategies inside Kubernetes clusters.

---

## ⚠️ Project Status

KubeRescue is currently under active development and testing.

This project is **not production-ready** and must **NOT** be deployed in production environments at this time.

Features, APIs, and behavior may change without notice as the architecture evolves.

Use only in development or staging clusters.

---

Modern Kubernetes environments require continuous monitoring and rapid remediation of failures such as:

- CrashLoopBackOff
- OOMKilled containers
- ImagePullBackOff
- Misconfigured deployments
- Restart storms

KubeRescue aims to implement:

```
Detection → Classification → Policy Evaluation → Safe Remediation → Observability
```

With SRE-grade safety controls such as:

- Retry limits
- Cooldown windows
- Idempotent actions
- Escalation policies
- Metrics export

---

### High-Level Controller Architecture

```mermaid
flowchart LR
    subgraph Kubernetes_Cluster
        A1[Pods]
        A2[Deployments]
        A3[Nodes]
    end

    subgraph KubeRescue_Controller
        B1[Watcher]
        B2[Analyzer]
        B3[Policy Engine]
        B4[Remediator]
        B5[Notifier]
        B6[Metrics Exporter]
    end

    A1 --> B1
    A2 --> B1
    A3 --> B1

    B1 --> B2
    B2 --> B3
    B3 --> B4
    B4 --> A1
    B4 --> A2

    B3 --> B5
    B4 --> B6
```

---

### Workflow Execution Flow

```mermaid
flowchart TD
    A[Start KubeRescue CLI] --> B[Initialize Kubernetes Client]
    B --> C[Load Policy Configuration YAML]
    C --> D[Start Cluster Watcher]

    D --> E{Failure Detected?}

    E -- No --> D
    E -- Yes --> F[Fetch Pod Details]
    F --> G[Collect Container Logs]

    G --> H[Analyzer Module]
    H --> I[Classify Failure Type]

    I --> J{Policy Exists?}

    J -- No --> K[Send Alert Only]
    J -- Yes --> L[Decision Engine]

    L --> M{Action Allowed? Retry Limit / Cooldown Check}

    M -- No --> N[Escalate Alert]
    M -- Yes --> O[Execute Remediation]

    O --> P{Action Type}

    P --> |Restart Pod| Q[Delete Pod]
    P --> |Scale Deployment| R[Update Replica Count]
    P --> |Rollback| S[Revert ReplicaSet]

    Q --> T[Verify Recovery]
    R --> T
    S --> T

    T --> U{Recovered?}

    U -- Yes --> V[Log Success + Metrics Update]
    U -- No --> W[Trigger Escalation Policy]

    V --> X[Send Slack / Webhook Notification]
    W --> X

    X --> D
```

---

### Current Features (Alpha)

- CrashLoopBackOff detection
- Automated pod restart
- Strict type checking (mypy)
- Security scanning (Bandit)
- Linting (Ruff)
- Code formatting (Black)
- Pre-commit enforcement
- CI pipeline ready

---

### Run CLI (Development Only)

```bash
kubrescue monitor --namespace default
```

---

### Engineering Principles

KubeRescue is built with the following principles:

- Safety over aggression
- Explicit remediation policies
- Idempotent operations
- Observable actions
- Strong typing and lint enforcement
- Security scanning by default

---

### Roadmap

- Retry & cooldown protection engine
- Kubernetes Watch API integration
- Policy-based YAML configuration
- Slack / webhook notifier
- Prometheus metrics exporter
- Helm chart deployment
- Docker image distribution
- Multi-namespace support

---

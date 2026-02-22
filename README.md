# KubeRescue

Autonomous Kubernetes failure detection and policy-driven auto-remediation engine for SRE teams.

### Controller Architecture Diagram

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

### Controller Architecture Diagram

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

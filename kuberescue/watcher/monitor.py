import time
from typing import Any
from kuberescue.config.k8s_client import load_kubernetes_client
from kuberescue.remediator.actions import restart_pod

CRASHLOOP_REASON = "CrashLoopBackOff"


def detect_crashloop(pod: Any) -> bool:
    """
    Detect if a pod is in CrashLoopBackOff state.
    """
    container_statuses = pod.status.container_statuses or []

    for container_status in container_statuses:
        state = container_status.state
        if state and state.waiting:
            if state.waiting.reason == CRASHLOOP_REASON:
                return True

    return False


def scan_namespace(
    v1: Any,
    namespace: str,
    dry_run: bool = False,
    max_restarts: int | None = None,
) -> tuple[int, int]:
    """
    Scan a namespace once and remediate CrashLoopBackOff pods.
    """
    pods = v1.list_namespaced_pod(namespace=namespace)
    detected = 0
    remediated = 0

    for pod in pods.items:
        if not detect_crashloop(pod):
            continue

        detected += 1
        pod_name = pod.metadata.name
        print(f"[KubeRescue] CrashLoop detected: {pod_name}")

        if max_restarts is not None and remediated >= max_restarts:
            print("[KubeRescue] Restart limit reached; " f"skipping pod: {pod_name}")
            continue

        if restart_pod(v1, pod_name, namespace, dry_run=dry_run):
            remediated += 1

    return detected, remediated


def monitor_namespace(
    namespace: str,
    interval_seconds: int = 10,
    once: bool = False,
    dry_run: bool = False,
    max_restarts: int | None = None,
) -> None:
    """
    Continuously monitor namespace for CrashLoopBackOff pods.
    """
    v1 = load_kubernetes_client()

    mode = "dry run" if dry_run else "active remediation"
    print(f"[KubeRescue] Monitoring namespace: {namespace} ({mode})")

    while True:
        detected, remediated = scan_namespace(
            v1,
            namespace,
            dry_run=dry_run,
            max_restarts=max_restarts,
        )
        print(
            "[KubeRescue] Scan complete: "
            f"detected={detected}, remediated={remediated}"
        )

        if once:
            return

        time.sleep(interval_seconds)

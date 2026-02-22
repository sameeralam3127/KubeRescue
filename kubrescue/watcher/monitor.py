import time
from typing import Any
from kubrescue.config.k8s_client import load_kubernetes_client
from kubrescue.remediator.actions import restart_pod


def detect_crashloop(pod: Any) -> bool:
    """
    Detect if a pod is in CrashLoopBackOff state.
    """
    container_statuses = pod.status.container_statuses or []

    for container_status in container_statuses:
        state = container_status.state
        if state and state.waiting:
            if state.waiting.reason == "CrashLoopBackOff":
                return True

    return False


def monitor_namespace(namespace: str) -> None:
    """
    Continuously monitor namespace for CrashLoopBackOff pods.
    """
    v1 = load_kubernetes_client()

    print(f"[KubeRescue] Monitoring namespace: {namespace}")

    while True:
        pods = v1.list_namespaced_pod(namespace=namespace)

        for pod in pods.items:
            if detect_crashloop(pod):
                pod_name = pod.metadata.name
                print(f"[KubeRescue] CrashLoop detected: {pod_name}")
                restart_pod(v1, pod_name, namespace)

        time.sleep(10)

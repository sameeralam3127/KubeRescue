from typing import Any


def restart_pod(v1: Any, pod_name: str, namespace: str) -> None:
    """
    Delete a pod to trigger restart by its controller.
    """
    try:
        v1.delete_namespaced_pod(
            name=pod_name,
            namespace=namespace,
        )
        print(f"[KubeRescue] Restarted pod: {pod_name}")
    except Exception as exc:
        print(f"[KubeRescue] Failed to restart {pod_name}: {exc}")

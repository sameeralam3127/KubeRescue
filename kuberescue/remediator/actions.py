from typing import Any


def restart_pod(v1: Any, pod_name: str, namespace: str, dry_run: bool = False) -> bool:
    """
    Delete a pod to trigger restart by its controller.
    """
    if dry_run:
        print(f"[KubeRescue] Dry run: would restart pod: {pod_name}")
        return True

    try:
        v1.delete_namespaced_pod(
            name=pod_name,
            namespace=namespace,
        )
        print(f"[KubeRescue] Restarted pod: {pod_name}")
        return True
    except Exception as exc:
        print(f"[KubeRescue] Failed to restart {pod_name}: {exc}")
        return False

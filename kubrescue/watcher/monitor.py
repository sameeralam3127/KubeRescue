import time
from kubernetes.client.rest import ApiException
from kubrescue.config.k8s_client import load_kubernetes_client
from kubrescue.remediator.actions import restart_pod


def monitor_namespace(namespace: str):
    v1 = load_kubernetes_client()

    print(f"[KubeRescue] Monitoring namespace: {namespace}")

    while True:
        try:
            pods = v1.list_namespaced_pod(namespace=namespace)

            for pod in pods.items:
                for container_status in pod.status.container_statuses or []:
                    state = container_status.state
                    if state and state.waiting:
                        reason = state.waiting.reason

                        if reason == "CrashLoopBackOff":
                            print(
                                f"[ALERT] CrashLoopBackOff detected in pod {pod.metadata.name}"
                            )
                            restart_pod(v1, pod.metadata.name, namespace)

        except ApiException as e:
            print(f"[ERROR] Kubernetes API error: {e}")

        time.sleep(10)
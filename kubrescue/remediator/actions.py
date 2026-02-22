def restart_pod(v1, pod_name: str, namespace: str):
    try:
        print(f"[ACTION] Restarting pod {pod_name}")
        v1.delete_namespaced_pod(name=pod_name, namespace=namespace)
        print(f"[SUCCESS] Pod {pod_name} deleted. Kubernetes will recreate it.")
    except Exception as e:
        print(f"[ERROR] Failed to restart pod {pod_name}: {e}")
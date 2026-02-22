from kubernetes import client, config


def load_kubernetes_client():
    try:
        # Try in-cluster config first
        config.load_incluster_config()
    except config.ConfigException:
        # Fallback to local kubeconfig
        config.load_kube_config()

    return client.CoreV1Api()
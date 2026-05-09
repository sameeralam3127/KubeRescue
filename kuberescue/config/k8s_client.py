from typing import Any
from kubernetes import client, config


def load_kubernetes_client() -> Any:
    """
    Load Kubernetes configuration.

    Tries in-cluster config first, then falls back to local kubeconfig.
    """
    try:
        config.load_incluster_config()
    except config.ConfigException:
        config.load_kube_config()

    return client.CoreV1Api()

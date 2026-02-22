from unittest.mock import MagicMock
from kubrescue.remediator.actions import restart_pod


def test_restart_pod_success() -> None:
    mock_v1 = MagicMock()

    restart_pod(mock_v1, "test-pod", "default")

    mock_v1.delete_namespaced_pod.assert_called_once_with(
        name="test-pod",
        namespace="default",
    )


def test_restart_pod_failure() -> None:
    mock_v1 = MagicMock()
    mock_v1.delete_namespaced_pod.side_effect = Exception("API error")

    restart_pod(mock_v1, "test-pod", "default")

    mock_v1.delete_namespaced_pod.assert_called_once()

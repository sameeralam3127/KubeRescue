from unittest.mock import MagicMock
from kuberescue.remediator.actions import restart_pod


def test_restart_pod_success() -> None:
    mock_v1 = MagicMock()

    result = restart_pod(mock_v1, "test-pod", "default")

    assert result is True
    mock_v1.delete_namespaced_pod.assert_called_once_with(
        name="test-pod",
        namespace="default",
    )


def test_restart_pod_failure() -> None:
    mock_v1 = MagicMock()
    mock_v1.delete_namespaced_pod.side_effect = Exception("API error")

    result = restart_pod(mock_v1, "test-pod", "default")

    assert result is False
    mock_v1.delete_namespaced_pod.assert_called_once()


def test_restart_pod_dry_run() -> None:
    mock_v1 = MagicMock()

    result = restart_pod(mock_v1, "test-pod", "default", dry_run=True)

    assert result is True
    mock_v1.delete_namespaced_pod.assert_not_called()

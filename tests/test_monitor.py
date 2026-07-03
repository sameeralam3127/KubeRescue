from typing import Any
from unittest.mock import MagicMock
from kuberescue.watcher.monitor import detect_crashloop, scan_namespace


def create_mock_pod(reason: str | None, name: str = "test-pod") -> Any:
    pod = MagicMock()

    container_status = MagicMock()
    state = MagicMock()
    waiting = MagicMock()

    waiting.reason = reason
    state.waiting = waiting
    container_status.state = state

    pod.status.container_statuses = [container_status]
    pod.metadata.name = name

    return pod


def test_detect_crashloop_true() -> None:
    pod = create_mock_pod("CrashLoopBackOff")
    assert detect_crashloop(pod) is True


def test_detect_crashloop_false() -> None:
    pod = create_mock_pod("ImagePullBackOff")
    assert detect_crashloop(pod) is False


def test_scan_namespace_restarts_crashloop_pods() -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = [
        create_mock_pod("CrashLoopBackOff", name="bad-pod"),
        create_mock_pod("Running", name="healthy-pod"),
    ]

    detected, remediated = scan_namespace(mock_v1, "default")

    assert detected == 1
    assert remediated == 1
    mock_v1.delete_namespaced_pod.assert_called_once_with(
        name="bad-pod",
        namespace="default",
    )


def test_scan_namespace_dry_run_does_not_delete_pods() -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = [
        create_mock_pod("CrashLoopBackOff", name="bad-pod"),
    ]

    detected, remediated = scan_namespace(mock_v1, "default", dry_run=True)

    assert detected == 1
    assert remediated == 1
    mock_v1.delete_namespaced_pod.assert_not_called()


def test_scan_namespace_respects_restart_limit() -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = [
        create_mock_pod("CrashLoopBackOff", name="bad-pod-1"),
        create_mock_pod("CrashLoopBackOff", name="bad-pod-2"),
    ]

    detected, remediated = scan_namespace(mock_v1, "default", max_restarts=1)

    assert detected == 2
    assert remediated == 1
    mock_v1.delete_namespaced_pod.assert_called_once_with(
        name="bad-pod-1",
        namespace="default",
    )

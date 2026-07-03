import json
from typing import Any
from unittest.mock import MagicMock
from kuberescue.watcher.monitor import (
    detect_crashloop,
    find_crashloop_containers,
    scan_namespace,
)


def create_mock_pod(
    reason: str | None,
    name: str = "test-pod",
    container_name: str = "app",
    restart_count: int = 3,
    last_reason: str | None = "Error",
    last_exit_code: int | None = 1,
) -> Any:
    pod = MagicMock()

    container_status = MagicMock()
    state = MagicMock()
    waiting = MagicMock()
    last_state = MagicMock()
    terminated = MagicMock()

    waiting.reason = reason
    state.waiting = waiting
    container_status.state = state
    container_status.name = container_name
    container_status.restart_count = restart_count

    terminated.reason = last_reason
    terminated.exit_code = last_exit_code
    last_state.terminated = terminated
    container_status.last_state = last_state

    pod.status.container_statuses = [container_status]
    pod.metadata.name = name

    return pod


def test_detect_crashloop_true() -> None:
    pod = create_mock_pod("CrashLoopBackOff")
    assert detect_crashloop(pod) is True


def test_detect_crashloop_false() -> None:
    pod = create_mock_pod("ImagePullBackOff")
    assert detect_crashloop(pod) is False


def test_find_crashloop_containers_includes_diagnostics() -> None:
    pod = create_mock_pod(
        "CrashLoopBackOff",
        name="bad-pod",
        container_name="worker",
        restart_count=7,
        last_reason="OOMKilled",
        last_exit_code=137,
    )

    findings = find_crashloop_containers(pod, "default")

    assert len(findings) == 1
    assert findings[0].pod == "bad-pod"
    assert findings[0].container == "worker"
    assert findings[0].restart_count == 7
    assert findings[0].last_terminated_reason == "OOMKilled"
    assert findings[0].last_exit_code == 137


def test_scan_namespace_restarts_crashloop_pods() -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = [
        create_mock_pod("CrashLoopBackOff", name="bad-pod"),
        create_mock_pod("Running", name="healthy-pod"),
    ]

    detected, remediated = scan_namespace(mock_v1, "default")

    assert detected == 1
    assert remediated == 1
    mock_v1.list_namespaced_pod.assert_called_once_with(
        namespace="default",
        label_selector=None,
    )
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


def test_scan_namespace_uses_label_selector() -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = []

    detected, remediated = scan_namespace(
        mock_v1,
        "default",
        selector="app=api",
    )

    assert detected == 0
    assert remediated == 0
    mock_v1.list_namespaced_pod.assert_called_once_with(
        namespace="default",
        label_selector="app=api",
    )


def test_scan_namespace_json_output(capsys: Any) -> None:
    mock_v1 = MagicMock()
    mock_v1.list_namespaced_pod.return_value.items = [
        create_mock_pod("CrashLoopBackOff", name="bad-pod"),
    ]

    detected, remediated = scan_namespace(
        mock_v1,
        "default",
        dry_run=True,
        output="json",
    )

    payload = json.loads(capsys.readouterr().out)

    assert detected == 1
    assert remediated == 1
    assert payload["detected"] == 1
    assert payload["remediated"] == 1
    assert payload["findings"][0]["pod"] == "bad-pod"
    mock_v1.delete_namespaced_pod.assert_not_called()

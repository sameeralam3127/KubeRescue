from typing import Any
from unittest.mock import MagicMock
from kubrescue.watcher.monitor import detect_crashloop


def create_mock_pod(reason: str | None) -> Any:
    pod = MagicMock()

    container_status = MagicMock()
    state = MagicMock()
    waiting = MagicMock()

    waiting.reason = reason
    state.waiting = waiting
    container_status.state = state

    pod.status.container_statuses = [container_status]

    return pod


def test_detect_crashloop_true() -> None:
    pod = create_mock_pod("CrashLoopBackOff")
    assert detect_crashloop(pod) is True


def test_detect_crashloop_false() -> None:
    pod = create_mock_pod("ImagePullBackOff")
    assert detect_crashloop(pod) is False

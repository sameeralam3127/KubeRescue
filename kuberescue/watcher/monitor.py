import json
import time
from dataclasses import asdict, dataclass
from typing import Any
from kuberescue.config.k8s_client import load_kubernetes_client
from kuberescue.remediator.actions import restart_pod

CRASHLOOP_REASON = "CrashLoopBackOff"


@dataclass(frozen=True)
class CrashLoopFinding:
    namespace: str
    pod: str
    container: str
    waiting_reason: str
    restart_count: int
    last_terminated_reason: str | None
    last_exit_code: int | None


def detect_crashloop(pod: Any) -> bool:
    """
    Detect if a pod is in CrashLoopBackOff state.
    """
    return bool(find_crashloop_containers(pod, namespace=""))


def find_crashloop_containers(pod: Any, namespace: str) -> list[CrashLoopFinding]:
    """
    Return diagnostic details for containers in CrashLoopBackOff.
    """
    container_statuses = getattr(pod.status, "container_statuses", None) or []
    pod_name = getattr(pod.metadata, "name", "<unknown>")
    findings = []

    for container_status in container_statuses:
        state = getattr(container_status, "state", None)
        waiting = getattr(state, "waiting", None) if state else None
        waiting_reason = getattr(waiting, "reason", None) if waiting else None

        if waiting_reason != CRASHLOOP_REASON:
            continue

        last_state = getattr(container_status, "last_state", None)
        terminated = getattr(last_state, "terminated", None) if last_state else None
        last_reason = getattr(terminated, "reason", None) if terminated else None
        last_exit_code = getattr(terminated, "exit_code", None) if terminated else None

        findings.append(
            CrashLoopFinding(
                namespace=namespace,
                pod=pod_name,
                container=getattr(container_status, "name", "<unknown>"),
                waiting_reason=waiting_reason,
                restart_count=getattr(container_status, "restart_count", 0),
                last_terminated_reason=last_reason,
                last_exit_code=last_exit_code,
            )
        )

    return findings


def format_finding(finding: CrashLoopFinding) -> str:
    """
    Format a finding for human-readable CLI output.
    """
    last_reason = finding.last_terminated_reason or "unknown"
    last_exit_code = (
        "unknown" if finding.last_exit_code is None else str(finding.last_exit_code)
    )
    return (
        f"pod={finding.pod} container={finding.container} "
        f"restarts={finding.restart_count} "
        f"last_reason={last_reason} last_exit_code={last_exit_code}"
    )


def scan_namespace(
    v1: Any,
    namespace: str,
    dry_run: bool = False,
    max_restarts: int | None = None,
    selector: str | None = None,
    output: str = "text",
) -> tuple[int, int]:
    """
    Scan a namespace once and remediate CrashLoopBackOff pods.
    """
    pods = v1.list_namespaced_pod(namespace=namespace, label_selector=selector)
    detected = 0
    remediated = 0
    findings_by_pod = []

    for pod in pods.items:
        findings = find_crashloop_containers(pod, namespace)
        if not findings:
            continue

        detected += 1
        pod_name = pod.metadata.name
        findings_by_pod.extend(findings)

        if output == "text":
            print(f"[KubeRescue] CrashLoop detected: {pod_name}")
            for finding in findings:
                print(f"[KubeRescue]   {format_finding(finding)}")

        if max_restarts is not None and remediated >= max_restarts:
            if output == "text":
                print(
                    "[KubeRescue] Restart limit reached; " f"skipping pod: {pod_name}"
                )
            continue

        if restart_pod(
            v1, pod_name, namespace, dry_run=dry_run, quiet=output == "json"
        ):
            remediated += 1

    if output == "json":
        print(
            json.dumps(
                {
                    "namespace": namespace,
                    "selector": selector,
                    "detected": detected,
                    "remediated": remediated,
                    "findings": [asdict(finding) for finding in findings_by_pod],
                },
                sort_keys=True,
            )
        )

    return detected, remediated


def monitor_namespace(
    namespace: str,
    interval_seconds: int = 10,
    once: bool = False,
    dry_run: bool = False,
    max_restarts: int | None = None,
    selector: str | None = None,
    output: str = "text",
) -> None:
    """
    Continuously monitor namespace for CrashLoopBackOff pods.
    """
    v1 = load_kubernetes_client()

    mode = "dry run" if dry_run else "active remediation"
    if output == "text":
        print(f"[KubeRescue] Monitoring namespace: {namespace} ({mode})")

    while True:
        detected, remediated = scan_namespace(
            v1,
            namespace,
            dry_run=dry_run,
            max_restarts=max_restarts,
            selector=selector,
            output=output,
        )
        if output == "text":
            print(
                "[KubeRescue] Scan complete: "
                f"detected={detected}, remediated={remediated}"
            )

        if once:
            return

        time.sleep(interval_seconds)

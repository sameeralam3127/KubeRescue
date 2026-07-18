package detect

import (
	corev1 "k8s.io/api/core/v1"
)

// ReasonCrashLoopBackOff is the container waiting reason set by the kubelet
// when a container keeps failing and is in restart backoff.
const ReasonCrashLoopBackOff = "CrashLoopBackOff"

// CrashLoop detects containers stuck in CrashLoopBackOff and records the
// restart count and last termination details as evidence.
type CrashLoop struct{}

// Name implements Detector.
func (CrashLoop) Name() string { return "crashloop" }

// Detect implements Detector.
func (CrashLoop) Detect(pod *corev1.Pod) []Finding {
	var findings []Finding
	ownerKind, ownerName := ControllerOf(pod)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting == nil || cs.State.Waiting.Reason != ReasonCrashLoopBackOff {
			continue
		}

		finding := Finding{
			Namespace:    pod.Namespace,
			Pod:          pod.Name,
			Container:    cs.Name,
			Reason:       ReasonCrashLoopBackOff,
			RestartCount: cs.RestartCount,
			OwnerKind:    ownerKind,
			OwnerName:    ownerName,
		}
		if term := cs.LastTerminationState.Terminated; term != nil {
			finding.LastTerminationReason = term.Reason
			exitCode := term.ExitCode
			finding.LastExitCode = &exitCode
		}
		findings = append(findings, finding)
	}
	return findings
}

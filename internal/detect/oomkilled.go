package detect

import (
	corev1 "k8s.io/api/core/v1"
)

// ReasonOOMKilled is the termination reason set by the kubelet when a
// container is killed for exceeding its memory limit.
const ReasonOOMKilled = "OOMKilled"

// OOMKilled detects containers whose most recent termination was caused by
// the out-of-memory killer.
type OOMKilled struct{}

// Name implements Detector.
func (OOMKilled) Name() string { return "oomkilled" }

// Detect implements Detector.
func (OOMKilled) Detect(pod *corev1.Pod) []Finding {
	var findings []Finding
	ownerKind, ownerName := ControllerOf(pod)

	for _, cs := range pod.Status.ContainerStatuses {
		term := cs.LastTerminationState.Terminated
		if term == nil || term.Reason != ReasonOOMKilled {
			continue
		}

		exitCode := term.ExitCode
		findings = append(findings, Finding{
			Namespace:             pod.Namespace,
			Pod:                   pod.Name,
			Container:             cs.Name,
			Reason:                ReasonOOMKilled,
			RestartCount:          cs.RestartCount,
			LastTerminationReason: term.Reason,
			LastExitCode:          &exitCode,
			OwnerKind:             ownerKind,
			OwnerName:             ownerName,
		})
	}
	return findings
}

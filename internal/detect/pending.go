package detect

import (
	corev1 "k8s.io/api/core/v1"
)

// ReasonFailedScheduling is used when a pending pod's PodScheduled
// condition reports the scheduler could not place it.
const ReasonFailedScheduling = "FailedScheduling"

// Pending detects pods the scheduler has not been able to place. Unlike the
// other detectors this is a pod-level condition rather than a per-container
// state, so it reports at most one finding per pod with an empty Container.
type Pending struct{}

// Name implements Detector.
func (Pending) Name() string { return "pending" }

// Detect implements Detector.
func (Pending) Detect(pod *corev1.Pod) []Finding {
	if pod.Status.Phase != corev1.PodPending {
		return nil
	}

	for _, cond := range pod.Status.Conditions {
		if cond.Type != corev1.PodScheduled || cond.Status != corev1.ConditionFalse {
			continue
		}

		ownerKind, ownerName := ControllerOf(pod)
		return []Finding{{
			Namespace: pod.Namespace,
			Pod:       pod.Name,
			Reason:    ReasonFailedScheduling,
			OwnerKind: ownerKind,
			OwnerName: ownerName,
			Message:   cond.Message,
		}}
	}
	return nil
}

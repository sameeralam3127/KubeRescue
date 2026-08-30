// Package detect identifies unhealthy workloads and collects the evidence
// needed to diagnose and remediate them.
//
// Detectors are pure functions over observed cluster state: they never talk
// to the API server and never mutate anything. This keeps them trivially
// testable and lets the engine decide how state is sourced (polling today,
// informers later).
package detect

import (
	corev1 "k8s.io/api/core/v1"
)

// Finding describes one unhealthy container together with the evidence
// collected about it. It is the unit of currency between detection,
// remediation, reporting, and (in later milestones) diagnosis and policy.
type Finding struct {
	Namespace             string `json:"namespace"`
	Pod                   string `json:"pod"`
	Container             string `json:"container"`
	Reason                string `json:"reason"`
	RestartCount          int32  `json:"restartCount"`
	LastTerminationReason string `json:"lastTerminationReason,omitempty"`
	LastExitCode          *int32 `json:"lastExitCode,omitempty"`
	OwnerKind             string `json:"ownerKind,omitempty"`
	OwnerName             string `json:"ownerName,omitempty"`
	// Message carries free-text detail the API server attaches to the
	// condition (a waiting reason or a scheduling condition), when the
	// detector has one worth surfacing.
	Message string `json:"message,omitempty"`
}

// Detector inspects a pod and reports zero or more findings.
type Detector interface {
	// Name identifies the detector in logs, reports, and configuration.
	Name() string
	// Detect returns one finding per affected container.
	Detect(pod *corev1.Pod) []Finding
}

// ControllerOf returns the kind and name of the controller that owns the
// pod, or empty strings when the pod has no controller.
func ControllerOf(pod *corev1.Pod) (kind, name string) {
	for _, ref := range pod.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind, ref.Name
		}
	}
	return "", ""
}

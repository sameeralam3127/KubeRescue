package detect

import (
	appsv1 "k8s.io/api/apps/v1"
)

// ReasonProgressDeadlineExceeded is the Progressing condition reason
// Kubernetes sets once a Deployment rollout has made no progress for
// spec.progressDeadlineSeconds.
const ReasonProgressDeadlineExceeded = "ProgressDeadlineExceeded"

// RolloutFinding describes a Deployment stuck mid-rollout. It is a distinct
// shape from Finding because a Deployment has no single container to blame
// and no restart count — overloading Finding's fields would misrepresent
// what was actually observed.
type RolloutFinding struct {
	Namespace           string `json:"namespace"`
	Deployment          string `json:"deployment"`
	Reason              string `json:"reason"`
	Message             string `json:"message,omitempty"`
	DesiredReplicas     int32  `json:"desiredReplicas"`
	UnavailableReplicas int32  `json:"unavailableReplicas"`
}

// RolloutDetector inspects a Deployment and reports zero or more findings.
type RolloutDetector interface {
	// Name identifies the detector in logs, reports, and configuration.
	Name() string
	// Detect returns findings for one Deployment.
	Detect(deploy *appsv1.Deployment) []RolloutFinding
}

// StuckRollout detects Deployments whose rollout has exceeded its progress
// deadline, using the Progressing condition Kubernetes already computes —
// no state needs to be tracked across scans.
type StuckRollout struct{}

// Name implements RolloutDetector.
func (StuckRollout) Name() string { return "stuck-rollout" }

// Detect implements RolloutDetector.
func (StuckRollout) Detect(deploy *appsv1.Deployment) []RolloutFinding {
	for _, cond := range deploy.Status.Conditions {
		if cond.Type != appsv1.DeploymentProgressing || cond.Reason != ReasonProgressDeadlineExceeded {
			continue
		}

		desired := int32(1)
		if deploy.Spec.Replicas != nil {
			desired = *deploy.Spec.Replicas
		}
		return []RolloutFinding{{
			Namespace:           deploy.Namespace,
			Deployment:          deploy.Name,
			Reason:              cond.Reason,
			Message:             cond.Message,
			DesiredReplicas:     desired,
			UnavailableReplicas: deploy.Status.UnavailableReplicas,
		}}
	}
	return nil
}

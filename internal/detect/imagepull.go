package detect

import (
	corev1 "k8s.io/api/core/v1"
)

// Waiting reasons the kubelet sets when it cannot pull a container image.
const (
	ReasonImagePullBackOff = "ImagePullBackOff"
	ReasonErrImagePull     = "ErrImagePull"
)

// ImagePull detects containers stuck unable to pull their image.
type ImagePull struct{}

// Name implements Detector.
func (ImagePull) Name() string { return "imagepull" }

// Detect implements Detector.
func (ImagePull) Detect(pod *corev1.Pod) []Finding {
	var findings []Finding
	ownerKind, ownerName := ControllerOf(pod)

	for _, cs := range pod.Status.ContainerStatuses {
		waiting := cs.State.Waiting
		if waiting == nil || (waiting.Reason != ReasonImagePullBackOff && waiting.Reason != ReasonErrImagePull) {
			continue
		}

		findings = append(findings, Finding{
			Namespace:    pod.Namespace,
			Pod:          pod.Name,
			Container:    cs.Name,
			Reason:       waiting.Reason,
			RestartCount: cs.RestartCount,
			OwnerKind:    ownerKind,
			OwnerName:    ownerName,
			Message:      waiting.Message,
		})
	}
	return findings
}

package detect

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func imagePullPod(name, reason, message string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       "Deployment",
				Name:       name + "-deploy",
				Controller: boolPtr(true),
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "app",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: reason, Message: message},
				},
			}},
		},
	}
}

func TestImagePullDetectsBackOffAndErr(t *testing.T) {
	for _, reason := range []string{ReasonImagePullBackOff, ReasonErrImagePull} {
		pod := imagePullPod("bad-image", reason, "manifest not found")
		findings := ImagePull{}.Detect(pod)
		if len(findings) != 1 {
			t.Fatalf("reason %s: expected 1 finding, got %d", reason, len(findings))
		}
		f := findings[0]
		if f.Reason != reason || f.Message != "manifest not found" {
			t.Errorf("reason %s: unexpected finding %+v", reason, f)
		}
		if f.OwnerKind != "Deployment" {
			t.Errorf("reason %s: owner = %s, want Deployment", reason, f.OwnerKind)
		}
	}
}

func TestImagePullIgnoresOtherWaitingReasons(t *testing.T) {
	if got := (ImagePull{}).Detect(healthyPod("ok")); len(got) != 0 {
		t.Errorf("healthy pod produced %d findings", len(got))
	}
	crashLoop := imagePullPod("looping", ReasonCrashLoopBackOff, "")
	if got := (ImagePull{}).Detect(crashLoop); len(got) != 0 {
		t.Errorf("CrashLoopBackOff pod produced %d imagepull findings", len(got))
	}
}

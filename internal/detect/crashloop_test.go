package detect

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func boolPtr(b bool) *bool { return &b }

func crashLoopPod(name string, opts ...func(*corev1.Pod)) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       "ReplicaSet",
				Name:       name + "-rs",
				Controller: boolPtr(true),
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "app",
				RestartCount: 7,
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: ReasonCrashLoopBackOff},
				},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						Reason:   "OOMKilled",
						ExitCode: 137,
					},
				},
			}},
		},
	}
	for _, opt := range opts {
		opt(pod)
	}
	return pod
}

func healthyPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "app",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			}},
		},
	}
}

func TestCrashLoopDetectsWaitingReason(t *testing.T) {
	findings := CrashLoop{}.Detect(crashLoopPod("bad-pod"))

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Pod != "bad-pod" || f.Container != "app" || f.Reason != ReasonCrashLoopBackOff {
		t.Errorf("unexpected finding identity: %+v", f)
	}
	if f.RestartCount != 7 {
		t.Errorf("restartCount = %d, want 7", f.RestartCount)
	}
	if f.LastTerminationReason != "OOMKilled" {
		t.Errorf("lastTerminationReason = %q, want OOMKilled", f.LastTerminationReason)
	}
	if f.LastExitCode == nil || *f.LastExitCode != 137 {
		t.Errorf("lastExitCode = %v, want 137", f.LastExitCode)
	}
	if f.OwnerKind != "ReplicaSet" || f.OwnerName != "bad-pod-rs" {
		t.Errorf("owner = %s/%s, want ReplicaSet/bad-pod-rs", f.OwnerKind, f.OwnerName)
	}
}

func TestCrashLoopIgnoresHealthyAndOtherWaitingReasons(t *testing.T) {
	if got := (CrashLoop{}).Detect(healthyPod("ok")); len(got) != 0 {
		t.Errorf("healthy pod produced %d findings", len(got))
	}

	imagePull := crashLoopPod("pulling", func(p *corev1.Pod) {
		p.Status.ContainerStatuses[0].State.Waiting.Reason = "ImagePullBackOff"
	})
	if got := (CrashLoop{}).Detect(imagePull); len(got) != 0 {
		t.Errorf("ImagePullBackOff pod produced %d crashloop findings", len(got))
	}
}

func TestControllerOfBarePod(t *testing.T) {
	pod := crashLoopPod("bare", func(p *corev1.Pod) {
		p.OwnerReferences = nil
	})
	if kind, name := ControllerOf(pod); kind != "" || name != "" {
		t.Errorf("ControllerOf(bare) = %s/%s, want empty", kind, name)
	}
}

package detect

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func oomKilledPod(name string) *corev1.Pod {
	return &corev1.Pod{
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
				RestartCount: 3,
				State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						Reason:   ReasonOOMKilled,
						ExitCode: 137,
					},
				},
			}},
		},
	}
}

func TestOOMKilledDetectsLastTerminationReason(t *testing.T) {
	findings := OOMKilled{}.Detect(oomKilledPod("mem-hog"))

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Reason != ReasonOOMKilled || f.Container != "app" {
		t.Errorf("unexpected finding identity: %+v", f)
	}
	if f.LastExitCode == nil || *f.LastExitCode != 137 {
		t.Errorf("lastExitCode = %v, want 137", f.LastExitCode)
	}
	if f.OwnerKind != "ReplicaSet" || f.OwnerName != "mem-hog-rs" {
		t.Errorf("owner = %s/%s, want ReplicaSet/mem-hog-rs", f.OwnerKind, f.OwnerName)
	}
}

func TestOOMKilledIgnoresOtherTerminationReasons(t *testing.T) {
	if got := (OOMKilled{}).Detect(healthyPod("ok")); len(got) != 0 {
		t.Errorf("healthy pod produced %d findings", len(got))
	}

	other := oomKilledPod("crashed")
	other.Status.ContainerStatuses[0].LastTerminationState.Terminated.Reason = "Error"
	if got := (OOMKilled{}).Detect(other); len(got) != 0 {
		t.Errorf("non-OOM termination produced %d findings", len(got))
	}
}

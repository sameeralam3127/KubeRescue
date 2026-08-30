package detect

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func unschedulablePod(name, message string) *corev1.Pod {
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
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{{
				Type:    corev1.PodScheduled,
				Status:  corev1.ConditionFalse,
				Reason:  "Unschedulable",
				Message: message,
			}},
		},
	}
}

func TestPendingDetectsUnschedulable(t *testing.T) {
	findings := Pending{}.Detect(unschedulablePod("stuck", "0/3 nodes are available: insufficient cpu"))

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Reason != ReasonFailedScheduling || f.Message != "0/3 nodes are available: insufficient cpu" {
		t.Errorf("unexpected finding: %+v", f)
	}
	if f.OwnerKind != "ReplicaSet" || f.OwnerName != "stuck-rs" {
		t.Errorf("owner = %s/%s, want ReplicaSet/stuck-rs", f.OwnerKind, f.OwnerName)
	}
}

func TestPendingIgnoresScheduledOrRunningPods(t *testing.T) {
	if got := (Pending{}).Detect(healthyPod("ok")); len(got) != 0 {
		t.Errorf("running pod produced %d findings", len(got))
	}

	scheduled := unschedulablePod("about-to-run", "")
	scheduled.Status.Conditions[0].Status = corev1.ConditionTrue
	if got := (Pending{}).Detect(scheduled); len(got) != 0 {
		t.Errorf("scheduled-true pod produced %d findings", len(got))
	}
}

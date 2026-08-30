package detect

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int32Ptr(i int32) *int32 { return &i }

func stuckDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
		Status: appsv1.DeploymentStatus{
			UnavailableReplicas: 2,
			Conditions: []appsv1.DeploymentCondition{{
				Type:    appsv1.DeploymentProgressing,
				Status:  "False",
				Reason:  ReasonProgressDeadlineExceeded,
				Message: "ReplicaSet has timed out progressing",
			}},
		},
	}
}

func TestStuckRolloutDetectsProgressDeadlineExceeded(t *testing.T) {
	findings := StuckRollout{}.Detect(stuckDeployment("api"))

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Deployment != "api" || f.Reason != ReasonProgressDeadlineExceeded {
		t.Errorf("unexpected finding: %+v", f)
	}
	if f.DesiredReplicas != 3 || f.UnavailableReplicas != 2 {
		t.Errorf("replicas = desired:%d unavailable:%d, want 3/2", f.DesiredReplicas, f.UnavailableReplicas)
	}
}

func TestStuckRolloutIgnoresHealthyRollouts(t *testing.T) {
	healthy := stuckDeployment("api")
	healthy.Status.Conditions[0].Reason = "NewReplicaSetAvailable"
	if got := (StuckRollout{}).Detect(healthy); len(got) != 0 {
		t.Errorf("healthy deployment produced %d findings", len(got))
	}

	noConditions := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "fresh", Namespace: "default"}}
	if got := (StuckRollout{}).Detect(noConditions); len(got) != 0 {
		t.Errorf("deployment with no conditions produced %d findings", len(got))
	}
}

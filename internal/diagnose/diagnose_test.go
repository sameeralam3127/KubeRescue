package diagnose

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/sameeralam3127/kuberescue/internal/detect"
)

func boolPtr(b bool) *bool    { return &b }
func int32Ptr(i int32) *int32 { return &i }

func oomPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "ReplicaSet", Name: name + "-rs", Controller: boolPtr(true),
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "app",
				RestartCount: 4,
				State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{Reason: detect.ReasonOOMKilled, ExitCode: 137},
				},
			}},
		},
	}
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

func stuckDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
		Status: appsv1.DeploymentStatus{
			UnavailableReplicas: 1,
			Conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentProgressing,
				Reason: detect.ReasonProgressDeadlineExceeded,
			}},
		},
	}
}

func podEvent(name, reason string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Name: name + "-evt", Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: name, Namespace: "default"},
		Reason:         reason,
		Message:        "container back-off restarting",
	}
}

func detectors() []detect.Detector {
	return []detect.Detector{detect.CrashLoop{}, detect.OOMKilled{}, detect.ImagePull{}, detect.Pending{}}
}

func TestRunCollectsPodFindingsAndEvents(t *testing.T) {
	client := fake.NewClientset(oomPod("mem-hog"), healthyPod("ok"), podEvent("mem-hog", "BackOff"))

	report, err := Run(context.Background(), client, detectors(), []detect.RolloutDetector{detect.StuckRollout{}}, Options{Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	if report.Detected != 1 || len(report.PodFindings) != 1 {
		t.Fatalf("detected = %d, podFindings = %d, want 1/1", report.Detected, len(report.PodFindings))
	}
	bundle := report.PodFindings[0]
	if bundle.Finding.Reason != detect.ReasonOOMKilled {
		t.Errorf("finding reason = %q, want OOMKilled", bundle.Finding.Reason)
	}
	if bundle.Explanation == "" {
		t.Error("expected a non-empty explanation")
	}
	if len(bundle.Events) != 1 || bundle.Events[0].Reason != "BackOff" {
		t.Errorf("events = %+v, want one BackOff event", bundle.Events)
	}
}

func TestRunCollectsRolloutFindings(t *testing.T) {
	client := fake.NewClientset(stuckDeployment("api"))

	report, err := Run(context.Background(), client, detectors(), []detect.RolloutDetector{detect.StuckRollout{}}, Options{Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	if report.Detected != 1 || len(report.RolloutFindings) != 1 {
		t.Fatalf("detected = %d, rolloutFindings = %d, want 1/1", report.Detected, len(report.RolloutFindings))
	}
	if report.RolloutFindings[0].Finding.Deployment != "api" {
		t.Errorf("deployment = %q, want api", report.RolloutFindings[0].Finding.Deployment)
	}
}

func TestRunNarrowedToOnePodSkipsRollouts(t *testing.T) {
	client := fake.NewClientset(oomPod("mem-hog"), oomPod("other"), stuckDeployment("api"))

	report, err := Run(context.Background(), client, detectors(), []detect.RolloutDetector{detect.StuckRollout{}}, Options{Namespace: "default", Pod: "mem-hog"})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.PodFindings) != 1 || report.PodFindings[0].Finding.Pod != "mem-hog" {
		t.Fatalf("podFindings = %+v, want exactly mem-hog", report.PodFindings)
	}
	if len(report.RolloutFindings) != 0 {
		t.Errorf("expected rollout detection to be skipped, got %d findings", len(report.RolloutFindings))
	}
}

func TestRunReturnsCleanReportWhenNothingUnhealthy(t *testing.T) {
	client := fake.NewClientset(healthyPod("ok"))

	report, err := Run(context.Background(), client, detectors(), []detect.RolloutDetector{detect.StuckRollout{}}, Options{Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Detected != 0 || len(report.PodFindings) != 0 || len(report.RolloutFindings) != 0 {
		t.Errorf("expected a clean report, got %+v", report)
	}
}

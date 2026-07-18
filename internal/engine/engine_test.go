package engine

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/remediate"
)

func boolPtr(b bool) *bool { return &b }

func crashLoopPod(name string) *corev1.Pod {
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
				RestartCount: 5,
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: detect.ReasonCrashLoopBackOff},
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

func newEngine(client kubernetes.Interface) *Engine {
	return &Engine{
		Client:    client,
		Detectors: []detect.Detector{detect.CrashLoop{}},
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestScanRestartsCrashLoopingPods(t *testing.T) {
	client := fake.NewClientset(crashLoopPod("bad-pod"), healthyPod("ok-pod"))

	report, err := newEngine(client).Scan(context.Background(), Options{Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	if report.Detected != 1 || report.Restarted != 1 || report.Failed != 0 {
		t.Errorf("report = detected:%d restarted:%d failed:%d, want 1/1/0",
			report.Detected, report.Restarted, report.Failed)
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "bad-pod", metav1.GetOptions{}); err == nil {
		t.Error("bad-pod was not deleted")
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "ok-pod", metav1.GetOptions{}); err != nil {
		t.Error("healthy pod was deleted")
	}
}

// Dry-run reports must never claim remediations happened. The Python
// predecessor counted dry-run actions as remediated; this pins the fix.
func TestScanDryRunCountsNothingAsRestarted(t *testing.T) {
	client := fake.NewClientset(crashLoopPod("bad-pod"))

	report, err := newEngine(client).Scan(context.Background(), Options{Namespace: "default", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}

	if report.Detected != 1 || report.Restarted != 0 {
		t.Errorf("dry run report = detected:%d restarted:%d, want 1/0", report.Detected, report.Restarted)
	}
	if !report.DryRun {
		t.Error("report must be marked dryRun")
	}
	if len(report.Actions) != 1 || report.Actions[0].Outcome != remediate.OutcomeDryRun {
		t.Errorf("actions = %+v, want one dry-run outcome", report.Actions)
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "bad-pod", metav1.GetOptions{}); err != nil {
		t.Error("dry run deleted the pod")
	}
}

func TestScanHonorsRestartBudget(t *testing.T) {
	client := fake.NewClientset(crashLoopPod("bad-1"), crashLoopPod("bad-2"), crashLoopPod("bad-3"))

	report, err := newEngine(client).Scan(context.Background(), Options{Namespace: "default", MaxRestarts: 1})
	if err != nil {
		t.Fatal(err)
	}

	if report.Detected != 3 {
		t.Errorf("detected = %d, want 3", report.Detected)
	}
	if report.Restarted != 1 {
		t.Errorf("restarted = %d, want 1", report.Restarted)
	}
	if report.Skipped != 2 {
		t.Errorf("skipped = %d, want 2", report.Skipped)
	}
}

func TestScanPassesLabelSelector(t *testing.T) {
	client := fake.NewClientset()
	var seenSelector string
	client.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		seenSelector = action.(k8stesting.ListAction).GetListRestrictions().Labels.String()
		return false, nil, nil
	})

	if _, err := newEngine(client).Scan(context.Background(), Options{Namespace: "default", Selector: "app=api"}); err != nil {
		t.Fatal(err)
	}
	if seenSelector != "app=api" {
		t.Errorf("selector sent to API = %q, want app=api", seenSelector)
	}
}

func TestScanReturnsErrorInsteadOfPanicking(t *testing.T) {
	client := fake.NewClientset()
	client.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("apiserver unavailable")
	})

	_, err := newEngine(client).Scan(context.Background(), Options{Namespace: "default"})
	if err == nil {
		t.Fatal("expected error from failed list")
	}
}

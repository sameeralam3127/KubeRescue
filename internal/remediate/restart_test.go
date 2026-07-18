package remediate

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func boolPtr(b bool) *bool { return &b }

func managedPod(name string) *corev1.Pod {
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
	}
}

func TestRestartPodDeletesManagedPod(t *testing.T) {
	pod := managedPod("bad-pod")
	client := fake.NewClientset(pod)

	result := RestartPod(context.Background(), client, pod, false)

	if result.Outcome != OutcomeRestarted {
		t.Fatalf("outcome = %s, want %s (reason: %s)", result.Outcome, OutcomeRestarted, result.Reason)
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "bad-pod", metav1.GetOptions{}); err == nil {
		t.Error("pod still exists after restart")
	}
}

func TestRestartPodSkipsBarePod(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "bare", Namespace: "default"}}
	client := fake.NewClientset(pod)

	result := RestartPod(context.Background(), client, pod, false)

	if result.Outcome != OutcomeSkipped {
		t.Fatalf("outcome = %s, want %s", result.Outcome, OutcomeSkipped)
	}
	if result.Reason == "" {
		t.Error("skip result must carry a reason")
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "bare", metav1.GetOptions{}); err != nil {
		t.Error("bare pod was deleted; it must never be")
	}
}

func TestRestartPodDryRunDoesNotDelete(t *testing.T) {
	pod := managedPod("bad-pod")
	client := fake.NewClientset(pod)

	result := RestartPod(context.Background(), client, pod, true)

	if result.Outcome != OutcomeDryRun {
		t.Fatalf("outcome = %s, want %s", result.Outcome, OutcomeDryRun)
	}
	if _, err := client.CoreV1().Pods("default").Get(context.Background(), "bad-pod", metav1.GetOptions{}); err != nil {
		t.Error("dry run deleted the pod")
	}
}

func TestRestartPodReportsAPIFailure(t *testing.T) {
	pod := managedPod("bad-pod")
	client := fake.NewClientset(pod)
	client.PrependReactor("delete", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("boom")
	})

	result := RestartPod(context.Background(), client, pod, false)

	if result.Outcome != OutcomeFailed {
		t.Fatalf("outcome = %s, want %s", result.Outcome, OutcomeFailed)
	}
	if result.Reason == "" {
		t.Error("failure result must carry the error")
	}
}

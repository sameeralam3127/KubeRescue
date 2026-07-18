// Package remediate executes corrective actions against the cluster.
//
// Every action reports a truthful Outcome: a dry run is never counted as a
// remediation, and a skipped pod always carries the reason it was skipped.
package remediate

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sameeralam3127/kuberescue/internal/detect"
)

// Outcome states what actually happened to a pod.
type Outcome string

const (
	// OutcomeRestarted means the pod was deleted and its controller will
	// recreate it.
	OutcomeRestarted Outcome = "restarted"
	// OutcomeDryRun means the pod would have been restarted, but dry-run
	// mode prevented any change.
	OutcomeDryRun Outcome = "dry-run"
	// OutcomeSkipped means no action was taken; Result.Reason explains why.
	OutcomeSkipped Outcome = "skipped"
	// OutcomeFailed means the action was attempted and the API call failed.
	OutcomeFailed Outcome = "failed"
)

// Result records the outcome of one remediation attempt.
type Result struct {
	Pod     string  `json:"pod"`
	Outcome Outcome `json:"outcome"`
	Reason  string  `json:"reason,omitempty"`
}

// RestartPod deletes a controller-managed pod so its controller recreates it.
//
// Pods without a controller are skipped: deleting a bare pod removes the
// workload permanently instead of restarting it.
func RestartPod(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod, dryRun bool) Result {
	if kind, _ := detect.ControllerOf(pod); kind == "" {
		return Result{
			Pod:     pod.Name,
			Outcome: OutcomeSkipped,
			Reason:  "pod has no controller; deleting it would remove the workload, not restart it",
		}
	}

	if dryRun {
		return Result{Pod: pod.Name, Outcome: OutcomeDryRun}
	}

	if err := client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
		return Result{
			Pod:     pod.Name,
			Outcome: OutcomeFailed,
			Reason:  fmt.Sprintf("delete failed: %v", err),
		}
	}
	return Result{Pod: pod.Name, Outcome: OutcomeRestarted}
}

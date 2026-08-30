package diagnose

import (
	"fmt"

	"github.com/sameeralam3127/kuberescue/internal/detect"
)

// explainPod returns a deterministic, human-readable explanation of a
// finding's likely cause and a suggested next step. These are fixed
// heuristics, not a model call: diagnosis stays reproducible and auditable,
// and no AI sits anywhere near the action path.
func explainPod(f detect.Finding) string {
	switch f.Reason {
	case detect.ReasonCrashLoopBackOff:
		if f.LastTerminationReason == detect.ReasonOOMKilled {
			return fmt.Sprintf(
				"container %q is crash-looping because it was OOM-killed (exit code %s); "+
					"raise its memory limit or find the leak before restarting it again.",
				f.Container, exitCodeString(f.LastExitCode))
		}
		return fmt.Sprintf(
			"container %q has restarted %d times and is in backoff; check its logs for the "+
				"startup error (last termination reason: %s, exit code: %s).",
			f.Container, f.RestartCount, orUnknown(f.LastTerminationReason), exitCodeString(f.LastExitCode))
	case detect.ReasonOOMKilled:
		return fmt.Sprintf(
			"container %q was killed for exceeding its memory limit; raise the limit or "+
				"investigate a possible memory leak.",
			f.Container)
	case detect.ReasonImagePullBackOff, detect.ReasonErrImagePull:
		return fmt.Sprintf(
			"container %q cannot pull its image: %s. Check the image name and tag, and that "+
				"the cluster has credentials for the registry.",
			f.Container, orUnknown(f.Message))
	case detect.ReasonFailedScheduling:
		return fmt.Sprintf(
			"the scheduler cannot place this pod: %s.",
			orUnknown(f.Message))
	default:
		return fmt.Sprintf("unhealthy: %s", f.Reason)
	}
}

// explainRollout returns a human-readable explanation for a stuck rollout.
func explainRollout(f detect.RolloutFinding) string {
	return fmt.Sprintf(
		"deployment %q has not progressed and exceeded its progress deadline "+
			"(%d/%d replicas unavailable): %s. Check readiness probes and resource requests "+
			"on the new ReplicaSet's pods.",
		f.Deployment, f.UnavailableReplicas, f.DesiredReplicas, orUnknown(f.Message))
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

func exitCodeString(code *int32) string {
	if code == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d", *code)
}

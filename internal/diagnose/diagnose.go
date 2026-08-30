// Package diagnose collects evidence for unhealthy workloads and explains
// their likely cause. Unlike internal/engine, it never mutates the
// cluster — diagnosis is read-only by design; remediation stays a separate,
// explicit decision.
package diagnose

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sameeralam3127/kuberescue/internal/detect"
)

// SchemaVersion identifies the JSON report format. Bump on any breaking
// change to Report, EvidenceBundle, or RolloutEvidenceBundle.
const SchemaVersion = "v1alpha1"

// Options controls a diagnostic run.
type Options struct {
	Namespace string
	Selector  string
	// Pod, when set, narrows the run to a single named pod and skips
	// rollout detection (a Deployment-level concern that doesn't apply to
	// explaining one pod).
	Pod string
}

// Report is the result of one diagnostic run.
type Report struct {
	SchemaVersion   string                  `json:"schemaVersion"`
	Timestamp       time.Time               `json:"timestamp"`
	Namespace       string                  `json:"namespace"`
	Selector        string                  `json:"selector,omitempty"`
	Detected        int                     `json:"detected"`
	PodFindings     []EvidenceBundle        `json:"podFindings"`
	RolloutFindings []RolloutEvidenceBundle `json:"rolloutFindings"`
}

// Run detects unhealthy pods and (in namespace-wide mode) stuck rollouts,
// and attaches events and an explanation to each finding.
func Run(
	ctx context.Context,
	client kubernetes.Interface,
	detectors []detect.Detector,
	rolloutDetectors []detect.RolloutDetector,
	opts Options,
) (*Report, error) {
	report := &Report{
		SchemaVersion:   SchemaVersion,
		Timestamp:       time.Now().UTC(),
		Namespace:       opts.Namespace,
		Selector:        opts.Selector,
		PodFindings:     []EvidenceBundle{},
		RolloutFindings: []RolloutEvidenceBundle{},
	}

	if err := diagnosePods(ctx, client, detectors, opts, report); err != nil {
		return nil, err
	}
	if opts.Pod == "" {
		if err := diagnoseRollouts(ctx, client, rolloutDetectors, opts, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

func diagnosePods(ctx context.Context, client kubernetes.Interface, detectors []detect.Detector, opts Options, report *Report) error {
	pods, err := client.CoreV1().Pods(opts.Namespace).List(ctx, metav1.ListOptions{LabelSelector: opts.Selector})
	if err != nil {
		return fmt.Errorf("listing pods in %q: %w", opts.Namespace, err)
	}

	for i := range pods.Items {
		pod := &pods.Items[i]
		if opts.Pod != "" && pod.Name != opts.Pod {
			continue
		}

		var findings []detect.Finding
		for _, d := range detectors {
			findings = append(findings, d.Detect(pod)...)
		}
		if len(findings) == 0 {
			continue
		}

		events, err := collectEvents(ctx, client, pod.Namespace, "Pod", pod.Name)
		if err != nil {
			return err
		}
		for _, f := range findings {
			report.PodFindings = append(report.PodFindings, EvidenceBundle{
				Finding:     f,
				Events:      events,
				Explanation: explainPod(f),
			})
		}
		report.Detected += len(findings)
	}
	return nil
}

func diagnoseRollouts(ctx context.Context, client kubernetes.Interface, detectors []detect.RolloutDetector, opts Options, report *Report) error {
	deployments, err := client.AppsV1().Deployments(opts.Namespace).List(ctx, metav1.ListOptions{LabelSelector: opts.Selector})
	if err != nil {
		return fmt.Errorf("listing deployments in %q: %w", opts.Namespace, err)
	}

	for i := range deployments.Items {
		deploy := &deployments.Items[i]

		var findings []detect.RolloutFinding
		for _, d := range detectors {
			findings = append(findings, d.Detect(deploy)...)
		}
		if len(findings) == 0 {
			continue
		}

		events, err := collectEvents(ctx, client, deploy.Namespace, "Deployment", deploy.Name)
		if err != nil {
			return err
		}
		for _, f := range findings {
			report.RolloutFindings = append(report.RolloutFindings, RolloutEvidenceBundle{
				Finding:     f,
				Events:      events,
				Explanation: explainRollout(f),
			})
		}
		report.Detected += len(findings)
	}
	return nil
}

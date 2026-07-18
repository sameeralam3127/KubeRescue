// Package engine wires detection and remediation into scan and monitor
// loops. It owns error resilience: a transient API failure degrades one
// scan, never the process.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/remediate"
)

// SchemaVersion identifies the JSON report format. Bump on any breaking
// change to Report, Finding, or Result.
const SchemaVersion = "v1alpha1"

// Options controls a scan.
type Options struct {
	Namespace string
	Selector  string
	DryRun    bool
	// MaxRestarts caps remediations per scan; 0 means unlimited.
	MaxRestarts int
}

// Report is the result of one scan. Counters reflect what actually
// happened: dry-run actions are never counted as restarts.
type Report struct {
	SchemaVersion string             `json:"schemaVersion"`
	Timestamp     time.Time          `json:"timestamp"`
	Namespace     string             `json:"namespace"`
	Selector      string             `json:"selector,omitempty"`
	DryRun        bool               `json:"dryRun"`
	Detected      int                `json:"detected"`
	Restarted     int                `json:"restarted"`
	Skipped       int                `json:"skipped"`
	Failed        int                `json:"failed"`
	Findings      []detect.Finding   `json:"findings"`
	Actions       []remediate.Result `json:"actions"`
}

// Engine runs detectors over a namespace and applies remediation.
type Engine struct {
	Client    kubernetes.Interface
	Detectors []detect.Detector
	Logger    *slog.Logger
}

// Scan performs one detection and remediation pass.
func (e *Engine) Scan(ctx context.Context, opts Options) (*Report, error) {
	pods, err := e.Client.CoreV1().Pods(opts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: opts.Selector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pods in %q: %w", opts.Namespace, err)
	}

	report := &Report{
		SchemaVersion: SchemaVersion,
		Timestamp:     time.Now().UTC(),
		Namespace:     opts.Namespace,
		Selector:      opts.Selector,
		DryRun:        opts.DryRun,
		Findings:      []detect.Finding{},
		Actions:       []remediate.Result{},
	}

	for i := range pods.Items {
		pod := &pods.Items[i]

		var findings []detect.Finding
		for _, d := range e.Detectors {
			findings = append(findings, d.Detect(pod)...)
		}
		if len(findings) == 0 {
			continue
		}

		report.Detected++
		report.Findings = append(report.Findings, findings...)
		e.Logger.Info("unhealthy pod detected",
			"namespace", pod.Namespace,
			"pod", pod.Name,
			"reason", findings[0].Reason,
			"restarts", findings[0].RestartCount,
		)

		if opts.MaxRestarts > 0 && report.Restarted+report.Failed >= opts.MaxRestarts {
			result := remediate.Result{
				Pod:     pod.Name,
				Outcome: remediate.OutcomeSkipped,
				Reason:  fmt.Sprintf("restart budget of %d reached for this scan", opts.MaxRestarts),
			}
			report.Actions = append(report.Actions, result)
			report.Skipped++
			e.Logger.Info("skipping pod", "pod", pod.Name, "reason", result.Reason)
			continue
		}

		result := remediate.RestartPod(ctx, e.Client, pod, opts.DryRun)
		report.Actions = append(report.Actions, result)
		switch result.Outcome {
		case remediate.OutcomeRestarted:
			report.Restarted++
			e.Logger.Info("pod restarted", "pod", pod.Name)
		case remediate.OutcomeDryRun:
			e.Logger.Info("dry run: would restart pod", "pod", pod.Name)
		case remediate.OutcomeSkipped:
			report.Skipped++
			e.Logger.Warn("pod skipped", "pod", pod.Name, "reason", result.Reason)
		case remediate.OutcomeFailed:
			report.Failed++
			e.Logger.Error("restart failed", "pod", pod.Name, "reason", result.Reason)
		}
	}

	return report, nil
}

// Monitor runs Scan on an interval until ctx is cancelled. Scan errors are
// logged and retried with exponential backoff (capped at maxBackoff); they
// never terminate the loop. Each successful report is passed to publish.
func (e *Engine) Monitor(ctx context.Context, opts Options, interval time.Duration, publish func(*Report)) error {
	const maxBackoff = 2 * time.Minute
	delay := interval

	for {
		report, err := e.Scan(ctx, opts)
		switch {
		case err != nil && ctx.Err() != nil:
			return ctx.Err()
		case err != nil:
			e.Logger.Error("scan failed; will retry", "error", err, "retryIn", delay)
			delay = min(delay*2, maxBackoff)
		default:
			publish(report)
			delay = interval
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

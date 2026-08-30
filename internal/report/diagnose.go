package report

import (
	"encoding/json"
	"io"

	"github.com/sameeralam3127/kuberescue/internal/diagnose"
)

// DiagnoseJSON writes a diagnose report as a single JSON document.
func DiagnoseJSON(w io.Writer, r *diagnose.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// DiagnoseText writes a human-readable rendering of a diagnose report.
func DiagnoseText(w io.Writer, r *diagnose.Report) error {
	ew := &errWriter{w: w}

	if r.Detected == 0 {
		ew.printf("No unhealthy workloads found in namespace %q.\n", r.Namespace)
		return ew.err
	}

	for _, b := range r.PodFindings {
		f := b.Finding
		owner := "none"
		if f.OwnerKind != "" {
			owner = f.OwnerKind + "/" + f.OwnerName
		}
		ew.printf("%s  %s/%s", f.Reason, f.Namespace, f.Pod)
		if f.Container != "" {
			ew.printf(" container=%s", f.Container)
		}
		ew.printf(" owner=%s\n", owner)
		ew.printf("  %s\n", b.Explanation)
		for _, e := range b.Events {
			ew.printf("  event: %s x%d — %s\n", e.Reason, e.Count, e.Message)
		}
	}

	for _, b := range r.RolloutFindings {
		f := b.Finding
		ew.printf("%s  deployment/%s/%s\n", f.Reason, f.Namespace, f.Deployment)
		ew.printf("  %s\n", b.Explanation)
		for _, e := range b.Events {
			ew.printf("  event: %s x%d — %s\n", e.Reason, e.Count, e.Message)
		}
	}

	ew.printf("\nSummary: detected=%d\n", r.Detected)
	return ew.err
}

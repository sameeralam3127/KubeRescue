// Package report renders scan reports for humans (text) and machines
// (JSON). Reports go to stdout; logs go to stderr — never mix the two.
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/sameeralam3127/kuberescue/internal/engine"
)

// JSON writes the report as a single JSON document.
func JSON(w io.Writer, r *engine.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// Text writes a human-readable summary of the report.
func Text(w io.Writer, r *engine.Report) error {
	ew := &errWriter{w: w}

	if r.Detected == 0 {
		ew.printf("No unhealthy pods found in namespace %q.\n", r.Namespace)
		return ew.err
	}

	actionByPod := make(map[string]string, len(r.Actions))
	for _, a := range r.Actions {
		s := string(a.Outcome)
		if a.Reason != "" {
			s += " (" + a.Reason + ")"
		}
		actionByPod[a.Pod] = s
	}

	seen := make(map[string]bool, len(r.Findings))
	for _, f := range r.Findings {
		if !seen[f.Pod] {
			seen[f.Pod] = true
			ew.printf("%s  %s/%s\n", f.Reason, f.Namespace, f.Pod)
		}
		exitCode := "unknown"
		if f.LastExitCode != nil {
			exitCode = fmt.Sprintf("%d", *f.LastExitCode)
		}
		lastReason := f.LastTerminationReason
		if lastReason == "" {
			lastReason = "unknown"
		}
		owner := "none"
		if f.OwnerKind != "" {
			owner = f.OwnerKind + "/" + f.OwnerName
		}
		ew.printf("  container=%s restarts=%d lastReason=%s exitCode=%s owner=%s\n",
			f.Container, f.RestartCount, lastReason, exitCode, owner)
		if action, ok := actionByPod[f.Pod]; ok {
			ew.printf("  action: %s\n", action)
		}
	}

	mode := ""
	if r.DryRun {
		mode = " (dry run — nothing was changed)"
	}
	ew.printf("\nSummary: detected=%d restarted=%d skipped=%d failed=%d%s\n",
		r.Detected, r.Restarted, r.Skipped, r.Failed, mode)
	return ew.err
}

// errWriter remembers the first write error and drops subsequent writes,
// so rendering code stays linear.
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

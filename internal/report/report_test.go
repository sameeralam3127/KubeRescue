package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/engine"
	"github.com/sameeralam3127/kuberescue/internal/remediate"
)

func sampleReport() *engine.Report {
	exitCode := int32(137)
	return &engine.Report{
		SchemaVersion: engine.SchemaVersion,
		Timestamp:     time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
		Namespace:     "default",
		DryRun:        true,
		Detected:      1,
		Findings: []detect.Finding{{
			Namespace:             "default",
			Pod:                   "api-1",
			Container:             "api",
			Reason:                detect.ReasonCrashLoopBackOff,
			RestartCount:          7,
			LastTerminationReason: "OOMKilled",
			LastExitCode:          &exitCode,
			OwnerKind:             "ReplicaSet",
			OwnerName:             "api-rs",
		}},
		Actions: []remediate.Result{{Pod: "api-1", Outcome: remediate.OutcomeDryRun}},
	}
}

func TestJSONIsValidAndVersioned(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sampleReport()); err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["schemaVersion"] != engine.SchemaVersion {
		t.Errorf("schemaVersion = %v, want %s", decoded["schemaVersion"], engine.SchemaVersion)
	}
	if decoded["restarted"] != float64(0) {
		t.Errorf("dry-run report claims restarted = %v", decoded["restarted"])
	}
}

func TestTextIncludesEvidenceAndDryRunNotice(t *testing.T) {
	var buf bytes.Buffer
	if err := Text(&buf, sampleReport()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{
		"CrashLoopBackOff", "default/api-1", "restarts=7", "OOMKilled",
		"exitCode=137", "ReplicaSet/api-rs", "dry run",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q:\n%s", want, out)
		}
	}
}

func TestTextCleanReport(t *testing.T) {
	var buf bytes.Buffer
	r := &engine.Report{Namespace: "default"}
	if err := Text(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No unhealthy pods") {
		t.Errorf("unexpected clean-report output: %s", buf.String())
	}
}

package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sameeralam3127/kuberescue/internal/detect"
	"github.com/sameeralam3127/kuberescue/internal/diagnose"
)

func sampleDiagnoseReport() *diagnose.Report {
	return &diagnose.Report{
		SchemaVersion: diagnose.SchemaVersion,
		Timestamp:     time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
		Namespace:     "default",
		Detected:      2,
		PodFindings: []diagnose.EvidenceBundle{{
			Finding: detect.Finding{
				Namespace: "default", Pod: "api-1", Container: "api",
				Reason: detect.ReasonOOMKilled, OwnerKind: "ReplicaSet", OwnerName: "api-rs",
			},
			Events:      []diagnose.EventInfo{{Reason: "BackOff", Message: "back-off restarting", Count: 3}},
			Explanation: "container was OOM-killed",
		}},
		RolloutFindings: []diagnose.RolloutEvidenceBundle{{
			Finding: detect.RolloutFinding{
				Namespace: "default", Deployment: "web", Reason: detect.ReasonProgressDeadlineExceeded,
			},
			Explanation: "deployment has not progressed",
		}},
	}
}

func TestDiagnoseJSONIsValidAndVersioned(t *testing.T) {
	var buf bytes.Buffer
	if err := DiagnoseJSON(&buf, sampleDiagnoseReport()); err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["schemaVersion"] != diagnose.SchemaVersion {
		t.Errorf("schemaVersion = %v, want %s", decoded["schemaVersion"], diagnose.SchemaVersion)
	}
}

func TestDiagnoseTextIncludesEvidenceAndExplanation(t *testing.T) {
	var buf bytes.Buffer
	if err := DiagnoseText(&buf, sampleDiagnoseReport()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{
		"OOMKilled", "default/api-1", "container was OOM-killed",
		"BackOff x3", "deployment/default/web", "deployment has not progressed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q:\n%s", want, out)
		}
	}
}

func TestDiagnoseTextCleanReport(t *testing.T) {
	var buf bytes.Buffer
	r := &diagnose.Report{Namespace: "default"}
	if err := DiagnoseText(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No unhealthy workloads") {
		t.Errorf("unexpected clean-report output: %s", buf.String())
	}
}

package diagnose

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sameeralam3127/kuberescue/internal/detect"
)

// EventInfo summarizes one Kubernetes Event relevant to a finding.
type EventInfo struct {
	Reason   string    `json:"reason"`
	Message  string    `json:"message"`
	Count    int32     `json:"count"`
	LastSeen time.Time `json:"lastSeen"`
}

// EvidenceBundle pairs a pod/container finding with the events observed for
// it and a human-readable explanation of the likely cause.
type EvidenceBundle struct {
	Finding     detect.Finding `json:"finding"`
	Events      []EventInfo    `json:"events,omitempty"`
	Explanation string         `json:"explanation"`
}

// RolloutEvidenceBundle pairs a Deployment finding with its events and
// explanation.
type RolloutEvidenceBundle struct {
	Finding     detect.RolloutFinding `json:"finding"`
	Events      []EventInfo           `json:"events,omitempty"`
	Explanation string                `json:"explanation"`
}

// collectEvents fetches events involving the named object. Filtering is
// done client-side rather than via a field selector: fake clientsets used
// in tests don't index custom fields like involvedObject.name, and a
// namespace's event volume is small enough that this stays cheap against a
// real API server too.
func collectEvents(ctx context.Context, client kubernetes.Interface, namespace, kind, name string) ([]EventInfo, error) {
	list, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing events in %q: %w", namespace, err)
	}

	var events []EventInfo
	for _, e := range list.Items {
		if e.InvolvedObject.Kind != kind || e.InvolvedObject.Name != name {
			continue
		}
		lastSeen := e.LastTimestamp.Time
		if lastSeen.IsZero() {
			lastSeen = e.EventTime.Time
		}
		events = append(events, EventInfo{
			Reason:   e.Reason,
			Message:  e.Message,
			Count:    e.Count,
			LastSeen: lastSeen,
		})
	}
	return events, nil
}

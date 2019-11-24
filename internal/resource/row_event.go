package resource

import "k8s.io/apimachinery/pkg/watch"

// RowEvent represents a call for action after a resource reconciliation.
// Tracks whether a resource got added, deleted or updated.
type RowEvent struct {
	Action watch.EventType
	Fields Row
	Deltas Row
}

func newRowEvent(a watch.EventType, f, d Row) *RowEvent {
	return &RowEvent{Action: a, Fields: f, Deltas: d}
}

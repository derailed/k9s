// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/describe"
)

// eventTimes uses the modern event fields when present, falling back to the
// legacy core/v1 fields. Missing timestamps must not become year-one dates.
func eventTimes(e v1.Event) (first, last time.Time, count int32) {
	first = e.EventTime.Time
	if first.IsZero() {
		first = e.FirstTimestamp.Time
	}
	count = e.Count
	if count < 1 {
		count = 1
	}
	if e.Series != nil {
		last, count = e.Series.LastObservedTime.Time, e.Series.Count
	} else if count > 1 {
		last = e.LastTimestamp.Time
	}
	return first.UTC(), last.UTC(), count
}

func formatDescribeEvents(events []v1.Event) string {
	if len(events) == 0 {
		return "Events: <none>\n"
	}
	// Sort a copy so formatting does not mutate the caller's list.
	events = append([]v1.Event(nil), events...)
	sort.SliceStable(events, func(i, j int) bool {
		a, _, _ := eventTimes(events[i])
		b, _, _ := eventTimes(events[j])
		return a.Before(b)
	})
	date, multipleDates := "", false
	for _, e := range events {
		first, last, _ := eventTimes(e)
		for _, timestamp := range []time.Time{first, last} {
			if timestamp.IsZero() {
				continue
			}
			d := timestamp.Format(time.DateOnly)
			if date == "" {
				date = d
			} else if date != d {
				multipleDates = true
			}
		}
	}
	layout := time.TimeOnly
	if multipleDates {
		layout = time.DateTime
	}
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return "<unknown>"
		}
		return t.Format(layout)
	}
	var buf bytes.Buffer
	tabs := tabwriter.NewWriter(&buf, 0, 8, 2, ' ', 0)
	w := describe.NewPrefixWriter(tabs)
	if date != "" && !multipleDates {
		w.Write(0, "Events (%s UTC):\n", date)
	} else {
		w.Write(0, "Events:\n")
	}
	w.Write(1, "Time ranges show first → last occurrence.\n")
	w.Write(1, "Time (UTC)\tCount\tType\tReason\tFrom\tMessage\n")
	w.Write(1, "----------\t-----\t----\t------\t----\t-------\n")
	for _, e := range events {
		first, last, count := eventTimes(e)
		interval := formatTime(first)
		if count > 1 {
			interval += " → " + formatTime(last)
		}
		source := e.Source.Component
		if source == "" {
			source = e.ReportingController
		}
		message := strings.TrimSpace(e.Message)
		if e.InvolvedObject.FieldPath != "" {
			message = fmt.Sprintf("%s: %s", e.InvolvedObject.FieldPath, message)
		}
		w.Write(1, "%s\t%d\t%s\t%s\t%s\t%s\n", interval, count, e.Type, e.Reason, source, message)
	}
	_ = tabs.Flush()
	return buf.String()
}

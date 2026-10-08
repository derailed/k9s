// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/describe"
)

var imageSizeRX = regexp.MustCompile(`Image size: ([0-9]+) bytes\.`)

// eventTimes uses the modern event fields when present, falling back to the
// legacy core/v1 fields. Missing timestamps must not become year-one dates.
func eventTimes(e v1.Event, location *time.Location) (first, last time.Time, count int32) {
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
	return first.In(location), last.In(location), count
}

func formatDescribeEvents(events []v1.Event) string {
	return formatDescribeEventsInLocation(events, time.Local)
}

func formatDescribeEventsInLocation(events []v1.Event, location *time.Location) string {
	if len(events) == 0 {
		return "Events: <none>\n"
	}
	// Sort a copy so formatting does not mutate the caller's list.
	events = append([]v1.Event(nil), events...)
	sort.SliceStable(events, func(i, j int) bool {
		a, _, _ := eventTimes(events[i], location)
		b, _, _ := eventTimes(events[j], location)
		return a.Before(b)
	})
	date, multipleDates := "", false
	offset, multipleOffsets := "", false
	for _, e := range events {
		first, last, _ := eventTimes(e, location)
		for _, timestamp := range []time.Time{first, last} {
			if timestamp.IsZero() {
				continue
			}
			zone := timestamp.Format("-0700")
			if offset == "" {
				offset = zone
			} else if offset != zone {
				multipleOffsets = true
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
	if multipleOffsets {
		layout += " -0700"
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
		w.Write(0, "Events (%s local):\n", date)
	} else {
		w.Write(0, "Events:\n")
	}
	w.Write(1, "Time ranges show first → last occurrence. Orange ⚠ rows indicate warnings.\n")
	w.Write(1, "  Time (local)\tReason\tFrom\tMessage\n")
	w.Write(1, "  ----------\t------\t----\t-------\n")
	for _, e := range events {
		first, last, count := eventTimes(e, location)
		interval := formatTime(first)
		if count > 1 {
			end := formatTime(last)
			if !multipleOffsets && !first.IsZero() && !last.IsZero() && first.Format(time.DateOnly) == last.Format(time.DateOnly) {
				end = last.Format(time.TimeOnly)
				if first.Hour() == last.Hour() {
					end = last.Format(":04:05")
					if first.Minute() == last.Minute() {
						end = last.Format("::05")
					}
				}
			}
			interval += fmt.Sprintf(" → %s (x%d)", end, count)
		}
		source := e.Source.Component
		if source == "" {
			source = e.ReportingController
		}
		message := readableImageSizes(strings.TrimSpace(e.Message))
		if e.InvolvedObject.FieldPath != "" {
			message = fmt.Sprintf("%s: %s", e.InvolvedObject.FieldPath, message)
		}
		marker := "  "
		if e.Type == v1.EventTypeWarning {
			marker = "⚠ "
		}
		// Mark every physical line so wrapped and multiline warnings keep their style.
		lines := strings.Split(message, "\n")
		w.Write(1, "%s%s\t%s\t%s\t%s\n", marker, interval, e.Reason, source, lines[0])
		for _, line := range lines[1:] {
			w.Write(1, "%s\t\t\t%s\n", marker, line)
		}
	}
	_ = tabs.Flush()
	return buf.String()
}

// readableImageSizes retains the exact bytes and adds a binary IEC size.
func readableImageSizes(message string) string {
	return imageSizeRX.ReplaceAllStringFunc(message, func(match string) string {
		digits := imageSizeRX.FindStringSubmatch(match)[1]
		size, err := strconv.ParseUint(digits, 10, 64)
		if err != nil {
			return match
		}
		value, unit := float64(size), "B"
		for _, next := range []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"} {
			if value < 1024 {
				break
			}
			value /= 1024
			unit = next
		}
		return fmt.Sprintf("Image size: %s bytes (%.2f %s).", digits, value, unit)
	})
}

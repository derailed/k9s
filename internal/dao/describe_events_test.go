// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFormatDescribeEvents(t *testing.T) {
	first := time.Date(2026, 10, 6, 15, 51, 33, 0, time.FixedZone("CEST", 7200))
	last := first.Add(4 * time.Second)
	tests := []struct {
		name  string
		event v1.Event
		want  string
	}{
		{"single", v1.Event{FirstTimestamp: metav1.NewTime(first)}, "13:51:33"},
		{"legacy repeat", v1.Event{FirstTimestamp: metav1.NewTime(first), LastTimestamp: metav1.NewTime(last), Count: 3}, "13:51:33 → ::37"},
		{"series", v1.Event{EventTime: metav1.NewMicroTime(first), FirstTimestamp: metav1.NewTime(first.Add(-time.Hour)), Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(last)}}, "13:51:33 → ::37"},
		{"missing", v1.Event{}, "<unknown>"},
		{"missing last", v1.Event{EventTime: metav1.NewMicroTime(first), Count: 3}, "13:51:33 → <unknown>"},
		{"cross midnight", v1.Event{EventTime: metav1.NewMicroTime(first), Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(first.Add(24 * time.Hour))}}, "2026-10-06 13:51:33 → 2026-10-07 13:51:33"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTestDescribeEvents([]v1.Event{tt.event})
			assert.Contains(t, got, tt.want)
			assert.Contains(t, got, "Time (local)")
			assert.NotContains(t, got, "Count")
			assert.NotContains(t, got, "Type")
			assert.NotContains(t, got, "0001-")
		})
	}
	assert.Equal(t, "Events: <none>\n", formatTestDescribeEvents(nil))
}

func TestFormatDescribeEventsOrderAndDetails(t *testing.T) {
	first := time.Date(2026, 10, 6, 13, 51, 33, 0, time.UTC)
	events := []v1.Event{
		{EventTime: metav1.NewMicroTime(first.Add(time.Hour)), Reason: "Later"},
		{EventTime: metav1.NewMicroTime(first), Reason: "Earlier", ReportingController: "kubelet", Message: " probe failed \n", InvolvedObject: v1.ObjectReference{FieldPath: "spec.containers{app}"}, Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(first.Add(4 * time.Second))}},
	}
	got := formatTestDescribeEvents(events)
	assert.Contains(t, got, "Events (2026-10-06 local):")
	assert.Regexp(t, `13:51:33 → ::37 \(x3\)\s+Earlier\s+kubelet\s+spec.containers\{app\}: probe failed`, got)
	assert.Regexp(t, `(?s)Earlier.*Later`, got)
	assert.Equal(t, "Later", events[0].Reason)
	events[0].EventTime = metav1.NewMicroTime(first.Add(24 * time.Hour))
	assert.Contains(t, formatTestDescribeEvents(events), "2026-10-07 13:51:33")
}

type eventTestConnection struct {
	client.Connection
	kube    kubernetes.Interface
	dynamic dynamic.Interface
}

func (c eventTestConnection) Dial() (kubernetes.Interface, error) { return c.kube, nil }
func (c eventTestConnection) DynDial() (dynamic.Interface, error) { return c.dynamic, nil }

func TestDescribeResourceEvents(t *testing.T) {
	for _, ns := range []string{"default", ""} {
		t.Run("namespace="+ns, func(t *testing.T) {
			gvr, kind := client.NewGVR("v1/pods"), "Pod"
			if ns == "" {
				gvr, kind = client.NewGVR("v1/nodes"), "Node"
			}
			obj := &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "v1", "kind": kind,
				"metadata": map[string]interface{}{"name": "example", "namespace": ns, "uid": "current-uid"},
			}}
			kube := fake.NewClientset()
			calls := 0
			kube.PrependReactor("list", "events", func(action ktesting.Action) (bool, runtime.Object, error) {
				calls++
				list := action.(ktesting.ListAction)
				assert.Equal(t, ns, action.GetNamespace())
				assert.True(t, list.GetListRestrictions().Fields.Matches(fields.Set{
					"involvedObject.name": "example", "involvedObject.namespace": ns,
					"involvedObject.kind": kind, "involvedObject.uid": "current-uid",
				}))
				assert.False(t, list.GetListRestrictions().Fields.Matches(fields.Set{
					"involvedObject.name": "example", "involvedObject.namespace": ns,
					"involvedObject.kind": kind, "involvedObject.uid": "old-uid",
				}))
				result := &v1.EventList{Items: []v1.Event{{Reason: "Started"}}}
				if calls == 1 {
					result.Continue = "next-page"
				} else {
					assert.Equal(t, "next-page", action.(interface{ GetListOptions() metav1.ListOptions }).GetListOptions().Continue)
				}
				return true, result, nil
			})
			conn := eventTestConnection{kube: kube, dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), obj)}
			got, err := describeResourceEvents(context.Background(), conn, gvr, ns, "example", kind)
			require.NoError(t, err)
			assert.Equal(t, 2, calls)
			assert.Equal(t, 2, strings.Count(got, "Started"))
		})
	}
}

func TestDescribeEventWarningsAndCounts(t *testing.T) {
	first := time.Date(2026, 10, 8, 7, 45, 44, 0, time.UTC)
	got := formatTestDescribeEvents([]v1.Event{
		{Type: v1.EventTypeWarning, Reason: "Unhealthy", EventTime: metav1.NewMicroTime(first), Series: &v1.EventSeries{Count: 24, LastObservedTime: metav1.NewMicroTime(first.Add(3 * time.Second))}, Message: "probe failed\nconnection refused"},
		{Type: v1.EventTypeNormal, Reason: "Pulled", EventTime: metav1.NewMicroTime(first), Message: "Image size: 373518519 bytes."},
	})
	assert.Contains(t, got, "Orange ⚠ rows indicate warnings.")
	assert.Contains(t, got, "⚠ 07:45:44 → ::47 (x24)")
	assert.Regexp(t, `⚠\s+connection refused`, got)
	assert.Contains(t, got, "Image size: 373518519 bytes (356.22 MiB).")
	assert.NotContains(t, got, "(x1)")
	assert.NotContains(t, got, "Normal")
	assert.NotContains(t, got, "Warning")
}

func TestReadableImageSizes(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"Image size: 373518519 bytes.", "Image size: 373518519 bytes (356.22 MiB)."},
		{"Image size: 0 bytes.", "Image size: 0 bytes (0.00 B)."},
		{"Image size: 1024 bytes.", "Image size: 1024 bytes (1.00 KiB)."},
		{"Image size: 1048576 bytes.", "Image size: 1048576 bytes (1.00 MiB)."},
		{"Image size: 999 bytes.", "Image size: 999 bytes (999.00 B)."},
		{"Image size: 1000 bytes.", "Image size: 1000 bytes (1000.00 B)."},
		{"Image size: 1500000000 bytes.", "Image size: 1500000000 bytes (1.40 GiB)."},
		{"Image size: 18446744073709551616 bytes.", "Image size: 18446744073709551616 bytes."},
		{"Container created", "Container created"},
	} {
		assert.Equal(t, tt.want, readableImageSizes(tt.input))
	}
}

func TestCompactEventRanges(t *testing.T) {
	first := time.Date(2026, 10, 8, 7, 36, 44, 0, time.UTC)
	for _, tt := range []struct {
		name string
		last time.Time
		want string
	}{
		{"same minute", first.Add(3 * time.Second), "07:36:44 → ::47 (x25)"},
		{"same hour", first.Add(71 * time.Second), "07:36:44 → :37:55 (x25)"},
		{"different hour", first.Add(time.Hour), "07:36:44 → 08:36:44 (x25)"},
		{"different day", first.Add(24 * time.Hour), "2026-10-08 07:36:44 → 2026-10-09 07:36:44 (x25)"},
		{"missing end", time.Time{}, "07:36:44 → <unknown> (x25)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTestDescribeEvents([]v1.Event{{EventTime: metav1.NewMicroTime(first), Series: &v1.EventSeries{Count: 25, LastObservedTime: metav1.NewMicroTime(tt.last)}}})
			assert.Contains(t, got, tt.want)
		})
	}
}

// Keep formatting expectations independent of the machine running the tests.
func formatTestDescribeEvents(events []v1.Event) string {
	return formatDescribeEventsInLocation(events, time.UTC)
}

func TestDescribeEventsLocalTimezone(t *testing.T) {
	location, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	first := time.Date(2026, 10, 8, 23, 45, 44, 0, time.UTC)
	events := []v1.Event{{EventTime: metav1.NewMicroTime(first), Count: 2, LastTimestamp: metav1.NewTime(first.Add(3 * time.Second))}}
	got := formatDescribeEventsInLocation(events, location)
	assert.Contains(t, got, "Events (2026-10-09 local):")
	assert.Contains(t, got, "01:45:44 → ::47 (x2)")
	assert.Contains(t, got, "Time (local)")
	assert.Equal(t, formatDescribeEventsInLocation(events, time.Local), formatDescribeEvents(events))
	// A repeated hour at the end of daylight saving must not look like a backwards range.
	events[0].EventTime = metav1.NewMicroTime(time.Date(2026, 10, 25, 0, 45, 0, 0, time.UTC))
	events[0].LastTimestamp = metav1.NewTime(time.Date(2026, 10, 25, 1, 15, 0, 0, time.UTC))
	assert.Contains(t, formatDescribeEventsInLocation(events, location), "02:45:00 +0200 → 02:15:00 +0100 (x2)")
}

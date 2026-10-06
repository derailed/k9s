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
		{"legacy repeat", v1.Event{FirstTimestamp: metav1.NewTime(first), LastTimestamp: metav1.NewTime(last), Count: 3}, "13:51:33 → 13:51:37"},
		{"series", v1.Event{EventTime: metav1.NewMicroTime(first), FirstTimestamp: metav1.NewTime(first.Add(-time.Hour)), Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(last)}}, "13:51:33 → 13:51:37"},
		{"missing", v1.Event{}, "<unknown>"},
		{"missing last", v1.Event{EventTime: metav1.NewMicroTime(first), Count: 3}, "13:51:33 → <unknown>"},
		{"cross midnight", v1.Event{EventTime: metav1.NewMicroTime(first), Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(first.Add(24 * time.Hour))}}, "2026-10-06 13:51:33 → 2026-10-07 13:51:33"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDescribeEvents([]v1.Event{tt.event})
			assert.Contains(t, got, tt.want)
			assert.Contains(t, got, "Time (UTC)")
			assert.Contains(t, got, "Count")
			assert.NotContains(t, got, "0001-")
		})
	}
	assert.Equal(t, "Events: <none>\n", formatDescribeEvents(nil))
}

func TestFormatDescribeEventsOrderAndDetails(t *testing.T) {
	first := time.Date(2026, 10, 6, 13, 51, 33, 0, time.UTC)
	events := []v1.Event{
		{EventTime: metav1.NewMicroTime(first.Add(time.Hour)), Reason: "Later"},
		{EventTime: metav1.NewMicroTime(first), Reason: "Earlier", ReportingController: "kubelet", Message: " probe failed \n", InvolvedObject: v1.ObjectReference{FieldPath: "spec.containers{app}"}, Series: &v1.EventSeries{Count: 3, LastObservedTime: metav1.NewMicroTime(first.Add(4 * time.Second))}},
	}
	got := formatDescribeEvents(events)
	assert.Contains(t, got, "Events (2026-10-06 UTC):")
	assert.Regexp(t, `13:51:33 → 13:51:37\s+3\s+Earlier\s+kubelet\s+spec.containers\{app\}: probe failed`, got)
	assert.Regexp(t, `(?s)Earlier.*Later`, got)
	assert.Equal(t, "Later", events[0].Reason)
	events[0].EventTime = metav1.NewMicroTime(first.Add(24 * time.Hour))
	assert.Contains(t, formatDescribeEvents(events), "2026-10-07 13:51:33")
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

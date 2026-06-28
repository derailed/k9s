// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cols(names ...string) []metav1.TableColumnDefinition {
	dd := make([]metav1.TableColumnDefinition, 0, len(names))
	for _, n := range names {
		dd = append(dd, metav1.TableColumnDefinition{Name: n})
	}
	return dd
}

func TestCellAt(t *testing.T) {
	defs := cols("Name", "Ready", "Status")
	r := &metav1.TableRow{Cells: []any{"p1", "1/1", "Running"}}

	uu := map[string]struct {
		col string
		e   any
	}{
		"present":     {col: "Name", e: "p1"},
		"present-mid": {col: "Ready", e: "1/1"},
		"missing":     {col: "Nope", e: nil},
	}
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, cellAt(r, defs, u.col))
		})
	}
}

func TestCellAtShortRow(t *testing.T) {
	// Column exists in defs but the row has fewer cells.
	defs := cols("Name", "Ready", "Desired")
	r := &metav1.TableRow{Cells: []any{"p1"}}
	assert.Nil(t, cellAt(r, defs, "Desired"))
}

func TestWorkloadStatus(t *testing.T) {
	uu := map[string]struct {
		gvr  *client.GVR
		defs []metav1.TableColumnDefinition
		row  *metav1.TableRow
		e    string
	}{
		"pod-running-ok": {
			gvr:  client.PodGVR,
			defs: cols("Name", "Ready", "Status"),
			row:  &metav1.TableRow{Cells: []any{"p1", "1/1", render.PhaseRunning}},
			e:    StatusOK,
		},
		"pod-completed-ok": {
			gvr:  client.PodGVR,
			defs: cols("Name", "Ready", "Status"),
			row:  &metav1.TableRow{Cells: []any{"p1", "0/1", render.PhaseCompleted}},
			e:    StatusOK,
		},
		"pod-not-ready-degraded": {
			gvr:  client.PodGVR,
			defs: cols("Name", "Ready", "Status"),
			row:  &metav1.TableRow{Cells: []any{"p1", "0/1", render.PhaseRunning}},
			e:    DegradedStatus,
		},
		"dp-ready-ok": {
			gvr:  client.DpGVR,
			defs: cols("Name", "Ready"),
			row:  &metav1.TableRow{Cells: []any{"d1", "2/2"}},
			e:    StatusOK,
		},
		"rs-degraded": {
			gvr:  client.RsGVR,
			defs: cols("Name", "Desired", "Ready"),
			row:  &metav1.TableRow{Cells: []any{"r1", int64(3), int64(1)}},
			e:    DegradedStatus,
		},
		// CRD with non-native printer columns must not panic and degrades to a
		// placeholder status rather than crashing.
		"crd-degrades-gracefully": {
			gvr:  client.NewGVR("examples.demo.io/v1/foos"),
			defs: cols("Name", "Phase", "Region"),
			row:  &metav1.TableRow{Cells: []any{"foo1", "Active", "us-east"}},
			e:    render.MissingValue,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, u.e, status(u.gvr, u.row, u.defs))
			})
		})
	}
}

func TestWorkloadReadiness(t *testing.T) {
	uu := map[string]struct {
		gvr  *client.GVR
		defs []metav1.TableColumnDefinition
		row  *metav1.TableRow
		e    string
	}{
		"pod": {
			gvr:  client.PodGVR,
			defs: cols("Name", "Ready"),
			row:  &metav1.TableRow{Cells: []any{"p1", "1/1"}},
			e:    "1/1",
		},
		"rs": {
			gvr:  client.RsGVR,
			defs: cols("Name", "Desired", "Ready"),
			row:  &metav1.TableRow{Cells: []any{"r1", int64(2), int64(2)}},
			e:    "2/2",
		},
		"svc-empty": {
			gvr:  client.SvcGVR,
			defs: cols("Name"),
			row:  &metav1.TableRow{Cells: []any{"s1"}},
			e:    "",
		},
		// CRD has no Ready column -> N/A, no panic.
		"crd-na": {
			gvr:  client.NewGVR("examples.demo.io/v1/foos"),
			defs: cols("Name", "Phase"),
			row:  &metav1.TableRow{Cells: []any{"foo1", "Active"}},
			e:    render.NAValue,
		},
		// Native GVR but printer columns missing -> N/A instead of panic.
		"pod-missing-col-na": {
			gvr:  client.PodGVR,
			defs: cols("Name"),
			row:  &metav1.TableRow{Cells: []any{"p1"}},
			e:    render.NAValue,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, u.e, readiness(u.gvr, u.row, u.defs))
			})
		})
	}
}

func TestWorkloadGVRs(t *testing.T) {
	uu := map[string]struct {
		ctx context.Context
		e   []*client.GVR
	}{
		"no-config-uses-defaults": {
			ctx: context.Background(),
			e:   resList,
		},
		"empty-config-uses-defaults": {
			ctx: context.WithValue(context.Background(), internal.KeyViewConfig, config.NewCustomView()),
			e:   resList,
		},
		"override": {
			ctx: func() context.Context {
				cv := config.NewCustomView()
				cv.Workloads[config.DefaultWorkloadGVRs] = []string{
					"v1/pods",
					"examples.demo.io/v1/foos",
				}
				return context.WithValue(context.Background(), internal.KeyViewConfig, cv)
			}(),
			e: []*client.GVR{
				client.PodGVR,
				client.NewGVR("examples.demo.io/v1/foos"),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, workloadGVRs(u.ctx))
		})
	}
}

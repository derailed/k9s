// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	k9sconfig "github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeWorkloadRes(gvr, ns, name, status, ready, valid string, obj runtime.Object) *render.WorkloadRes {
	return &render.WorkloadRes{
		Row: metav1.TableRow{
			Object: runtime.RawExtension{Object: obj},
			Cells: []any{
				gvr,
				ns,
				name,
				status,
				ready,
				valid,
				metav1.Time{Time: time.Now().Add(-5 * time.Minute)},
			},
		},
	}
}

func TestWorkloadRenderDefault(t *testing.T) {
	w := render.Workload{}
	r := model1.NewRow(7)

	res := makeWorkloadRes(client.DpGVR.String(), "default", "web", "OK", "2/2", "", nil)
	require.NoError(t, w.Render(res, "", &r))
	assert.Equal(t, "apps/v1/deployments|default|web", r.ID)
	assert.Equal(t, model1.Fields{"apps/v1/deployments", "default", "web", "OK", "2/2", ""}, r.Fields[:6])
	assert.Equal(t, []string{"KIND", "NAMESPACE", "NAME", "STATUS", "READY", "VALID", "AGE"}, w.Header("").ColumnNames(true))
}

func TestWorkloadCustomColumns(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"metadata": map[string]any{
				"name":      "web",
				"namespace": "default",
				"labels": map[string]any{
					"app": "frontend",
				},
			},
		},
	}

	w := render.Workload{}
	w.SetViewSetting(&k9sconfig.ViewSetting{
		Columns: []string{
			"KIND|H",
			"NAME",
			"READY",
			"STATUS|W",
			"APP:.metadata.labels.app",
			"AGE",
		},
	})

	h := w.Header("")
	require.GreaterOrEqual(t, len(h), 6)
	assert.Equal(t, "KIND", h[0].Name)
	assert.True(t, h[0].Hide)
	assert.Equal(t, "NAME", h[1].Name)
	assert.False(t, h[1].Hide)
	assert.Equal(t, "READY", h[2].Name)
	assert.Equal(t, "STATUS", h[3].Name)
	assert.True(t, h[3].Wide)
	assert.Equal(t, "APP", h[4].Name)
	assert.Equal(t, "AGE", h[5].Name)
	assert.True(t, h[5].Time)

	r := model1.NewRow(7)
	res := makeWorkloadRes(client.DpGVR.String(), "default", "web", "OK", "2/2", "", obj)
	require.NoError(t, w.Render(res, "", &r))
	assert.Equal(t, "apps/v1/deployments|default|web", r.ID)
	assert.Equal(t, "apps/v1/deployments", r.Fields[0])
	assert.Equal(t, "web", r.Fields[1])
	assert.Equal(t, "2/2", r.Fields[2])
	assert.Equal(t, "OK", r.Fields[3])
	assert.Equal(t, "frontend", r.Fields[4])
	assert.NotEmpty(t, r.Fields[5])
}

func TestWorkloadCustomColumnsNilObject(t *testing.T) {
	w := render.Workload{}
	w.SetViewSetting(&k9sconfig.ViewSetting{
		Columns: []string{
			"KIND|H",
			"NAME",
			"READY",
			"APP:.metadata.labels.app",
			"AGE",
		},
	})

	r := model1.NewRow(7)
	res := makeWorkloadRes(client.StsGVR.String(), "default", "db", "OK", "3/3", "", nil)
	require.NoError(t, w.Render(res, "", &r))
	assert.Equal(t, "apps/v1/statefulsets|default|db", r.ID)
	assert.Equal(t, "apps/v1/statefulsets", r.Fields[0])
	assert.Equal(t, "db", r.Fields[1])
	assert.Equal(t, "3/3", r.Fields[2])
	assert.Equal(t, render.NAValue, r.Fields[3])
	assert.NotEmpty(t, r.Fields[4])
}

func TestWorkloadColorer(t *testing.T) {
	prevErr := model1.ErrColor
	prevPending := model1.PendingColor
	prevStd := model1.StdColor
	t.Cleanup(func() {
		model1.ErrColor = prevErr
		model1.PendingColor = prevPending
		model1.StdColor = prevStd
	})
	model1.ErrColor = tcell.ColorRed
	model1.PendingColor = tcell.ColorYellow
	model1.StdColor = tcell.ColorWhite

	w := render.Workload{}
	defaultHeader := w.Header("")

	wCustom := render.Workload{}
	wCustom.SetViewSetting(&k9sconfig.ViewSetting{
		Columns: []string{"KIND|H", "NAME", "STATUS", "READY", "AGE"},
	})
	hiddenKindHeader := wCustom.Header("")

	uu := map[string]struct {
		header model1.Header
		re     model1.RowEvent
		e      tcell.Color
	}{
		"deploy-ready": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/deployments|default|web",
					Fields: model1.Fields{"apps/v1/deployments", "default", "web", "OK", "2/2", "", "5m"},
				},
			},
			e: model1.StdColor,
		},
		"deploy-not-ready": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/deployments|default|web",
					Fields: model1.Fields{"apps/v1/deployments", "default", "web", "DEGRADED", "1/2", "DEGRADED", "5m"},
				},
			},
			e: model1.ErrColor,
		},
		"deploy-scaled-to-zero": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/deployments|default|web",
					Fields: model1.Fields{"apps/v1/deployments", "default", "web", "OK", "0/0", "", "5m"},
				},
			},
			e: model1.PendingColor,
		},
		"sts-ready": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/statefulsets|default|db",
					Fields: model1.Fields{"apps/v1/statefulsets", "default", "db", "OK", "3/3", "", "5m"},
				},
			},
			e: model1.StdColor,
		},
		"sts-not-ready": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/statefulsets|default|db",
					Fields: model1.Fields{"apps/v1/statefulsets", "default", "db", "DEGRADED", "1/3", "DEGRADED", "5m"},
				},
			},
			e: model1.ErrColor,
		},
		"sts-scaled-to-zero": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/statefulsets|default|db",
					Fields: model1.Fields{"apps/v1/statefulsets", "default", "db", "OK", "0/0", "", "5m"},
				},
			},
			e: model1.StdColor,
		},
		"pod-degraded-keeps-pending": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "v1/pods|default|p1",
					Fields: model1.Fields{"v1/pods", "default", "p1", "DEGRADED", "0/1", "DEGRADED", "5m"},
				},
			},
			e: model1.PendingColor,
		},
		"daemonset-degraded-keeps-pending": {
			header: defaultHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/daemonsets|default|ds1",
					Fields: model1.Fields{"apps/v1/daemonsets", "default", "ds1", "DEGRADED", "0/1", "DEGRADED", "5m"},
				},
			},
			e: model1.PendingColor,
		},
		"deploy-scaled-to-zero-with-hidden-kind": {
			header: hiddenKindHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/deployments|default|web",
					Fields: model1.Fields{"apps/v1/deployments", "web", "OK", "0/0", "5m", "default", ""},
				},
			},
			e: model1.PendingColor,
		},
		"deploy-not-ready-with-hidden-kind": {
			header: hiddenKindHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/deployments|default|web",
					Fields: model1.Fields{"apps/v1/deployments", "web", "DEGRADED", "1/2", "5m", "default", "DEGRADED"},
				},
			},
			e: model1.ErrColor,
		},
		"sts-not-ready-with-hidden-kind": {
			header: hiddenKindHeader,
			re: model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					ID:     "apps/v1/statefulsets|default|db",
					Fields: model1.Fields{"apps/v1/statefulsets", "db", "DEGRADED", "1/3", "5m", "default", "DEGRADED"},
				},
			},
			e: model1.ErrColor,
		},
	}

	colorer := w.ColorerFunc()
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, colorer("", u.header, &u.re))
		})
	}
}

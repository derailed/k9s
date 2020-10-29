package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	render.AddColor = tcell.ColorBlue
	render.HighlightColor = tcell.ColorYellow
	render.CompletedColor = tcell.ColorGray
	render.StdColor = tcell.ColorWhite
	render.ErrColor = tcell.ColorRed
	render.KillColor = tcell.ColorGray
}

func TestPodColorer(t *testing.T) {
	stdHeader := render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME"},
		render.HeaderColumn{Name: "READY"},
		render.HeaderColumn{Name: "RESTARTS"},
		render.HeaderColumn{Name: "STATUS"},
		render.HeaderColumn{Name: "VALID"},
	}

	uu := map[string]struct {
		re render.RowEvent
		h  render.Header
		e  tcell.Color
	}{
		"valid": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Running, ""},
				},
			},
			e: render.StdColor,
		},
		"init": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, ""},
				},
			},
			e: render.AddColor,
		},
		"init-err": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, "blah"},
				},
			},
			e: render.AddColor,
		},
		"initialized": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Initialized, "blah"},
				},
			},
			e: render.HighlightColor,
		},
		"completed": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Completed, "blah"},
				},
			},
			e: render.CompletedColor,
		},
		"terminating": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Terminating, "blah"},
				},
			},
			e: render.KillColor,
		},
		"invalid": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "Running", "blah"},
				},
			},
			e: render.ErrColor,
		},
		"unknown-cool": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: render.AddColor,
		},
		"unknown-err": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", "doh"},
				},
			},
			e: render.ErrColor,
		},
		"status": {
			h: stdHeader[0:3],
			re: render.RowEvent{
				Kind: render.EventDelete,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: render.KillColor,
		},
	}

	var r render.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", u.h, u.re))
		})
	}
}

func TestPodRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po"),
		MX:  makePodMX("nginx", "100m", "50Mi"),
	}

	var po render.Pod
	r := render.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "●", "1/1", "0", "Running", "100", "50", "100:0", "70:170", "100", "0", "71", "29", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:17])
}

func BenchmarkPodRender(b *testing.B) {
	pom := render.PodWithMetrics{
		Raw: load(b, "po"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}
	var po render.Pod
	r := render.NewRow(12)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = po.Render(&pom, "", &r)
	}
}

func TestPodInitRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po_init"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}

	var po render.Pod
	r := render.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "●", "1/1", "0", "Init:0/1", "10", "10", "100:0", "70:170", "10", "0", "14", "5", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:17])
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodMX(name, cpu, mem string) *mv1beta1.PodMetrics {
	return &mv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []mv1beta1.ContainerMetrics{
			{Usage: makeRes(cpu, mem)},
		},
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := res.ParseQuantity(c)
	mem, _ := res.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}

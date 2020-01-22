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

type (
	colorerUC struct {
		ns string
		r  render.RowEvent
		e  tcell.Color
	}

	colorerUCs []colorerUC
)

func TestPodColorer(t *testing.T) {
	var (
		nsRow      = render.Row{Fields: render.Fields{"blee", "fred", "1/1", "Running"}}
		toastNS    = render.Row{Fields: render.Fields{"blee", "fred", "1/1", "Boom"}}
		notReadyNS = render.Row{Fields: render.Fields{"blee", "fred", "0/1", "Boom"}}
		row        = render.Row{Fields: render.Fields{"fred", "1/1", "Running"}}
		toast      = render.Row{Fields: render.Fields{"fred", "1/1", "Boom"}}
		notReady   = render.Row{Fields: render.Fields{"fred", "0/1", "Boom"}}
	)

	uu := colorerUCs{
		// Add allNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: nsRow}, render.AddColor},
		// Add Namespaced
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: row}, render.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: nsRow}, render.ModColor},
		// Mod Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: row}, render.ModColor},
		// Mod Busted AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: toastNS}, render.ErrColor},
		// Mod Busted Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: toast}, render.ErrColor},
		// NotReady AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: notReadyNS}, render.ErrColor},
		// NotReady Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: notReady}, render.ErrColor},
	}

	var p render.Pod
	f := p.ColorerFunc()
	for _, u := range uu {
		assert.Equal(t, u.e, f(u.ns, u.r))
	}
}

func TestPodRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}

	var po render.Pod
	r := render.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "1/1", "Running", "0", "10", "10", "10", "14", "0", "5", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:14])
}

func BenchmarkPodRender(b *testing.B) {
	pom := render.PodWithMetrics{
		Raw: load(b, "po"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}
	var po render.Pod
	r := render.NewRow(12)

	b.ResetTimer()
	b.ReportAllocs()
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
	e := render.Fields{"default", "nginx", "1/1", "Init:0/1", "0", "10", "10", "10", "14", "0", "5", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:14])
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

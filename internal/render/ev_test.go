package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestEventRender(t *testing.T) {
	c := render.Event{}
	r := render.NewRow(7)
	c.Render(load(t, "ev"), "", &r)

	assert.Equal(t, "default/hello-1567197780-mn4mv.15bfce150bd764dd", r.ID)
	assert.Equal(t, render.Fields{"default", "pod:hello-1567197780-mn4mv", "Normal", "Pulled", "kubelet", "1", `Successfully pulled image "blang/busybox-bash"`}, r.Fields[:7])
}

func BenchmarkEventRender(b *testing.B) {
	ev := load(b, "ev")
	var re render.Event
	r := render.NewRow(7)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = re.Render(&ev, "", &r)
	}
}

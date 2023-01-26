package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestServiceRender(t *testing.T) {
	c := render.Service{}
	r := render.NewRow(4)

	assert.NoError(t, c.Render(load(t, "svc"), "", &r))
	assert.Equal(t, "default/dictionary1", r.ID)
	assert.Equal(t, render.Fields{"default", "dictionary1", "ClusterIP", "10.47.248.116", "", "app=dictionary1", "http:4001►0"}, r.Fields[:7])
}

func BenchmarkSvcRender(b *testing.B) {
	var svc render.Service
	r := render.NewRow(4)
	s := load(b, "svc")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = svc.Render(s, "", &r)
	}
}

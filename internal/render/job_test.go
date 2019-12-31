package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestJobRender(t *testing.T) {
	c := render.Job{}
	r := render.NewRow(4)
	c.Render(load(t, "job"), "", &r)

	assert.Equal(t, "default/hello-1567179180", r.ID)
	assert.Equal(t, render.Fields{"default", "hello-1567179180", "1/1", "8s", "c1", "blang/busybox-bash"}, r.Fields[:6])
}

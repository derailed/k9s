package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestHorizontalPodAutoscalerRender(t *testing.T) {
	c := render.HorizontalPodAutoscaler{}
	r := render.NewRow(7)
	c.Render(load(t, "hpa"), "", &r)

	assert.Equal(t, "default/nginx", r.ID)
	assert.Equal(t, render.Fields{"default", "nginx", "nginx", "<unknown>/10%", "1", "10"}, r.Fields[:6])
}

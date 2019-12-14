package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestIngressRender(t *testing.T) {
	c := render.Ingress{}
	r := render.NewRow(6)
	c.Render(load(t, "ing"), "", &r)

	assert.Equal(t, "default/test-ingress", r.ID)
	assert.Equal(t, render.Fields{"default", "test-ingress", "*", "", "80"}, r.Fields[:5])
}

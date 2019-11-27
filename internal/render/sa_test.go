package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountRender(t *testing.T) {
	c := render.ServiceAccount{}
	r := render.NewRow(4)
	c.Render(load(t, "sa"), "", &r)

	assert.Equal(t, "default/blee", r.ID)
	assert.Equal(t, render.Fields{"default", "blee", "2"}, r.Fields[:3])
}

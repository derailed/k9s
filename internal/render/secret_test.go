package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestSecRender(t *testing.T) {
	c := render.Secret{}
	r := render.NewRow(4)

	c.Render(load(t, "sec"), "", &r)
	assert.Equal(t, "default/s1", r.ID)
	assert.Equal(t, render.Fields{"default", "s1", "Opaque", "2"}, r.Fields[:4])
}

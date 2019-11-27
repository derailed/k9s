package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceRender(t *testing.T) {
	c := render.Namespace{}
	r := render.NewRow(3)
	c.Render(load(t, "ns"), "-", &r)

	assert.Equal(t, "kube-system", r.ID)
	assert.Equal(t, render.Fields{"kube-system", "Active"}, r.Fields[:2])
}

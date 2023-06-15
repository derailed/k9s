package render_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derailed/k9s/internal/render"
)

func TestStorageClassRender(t *testing.T) {
	c := render.StorageClass{}
	r := render.NewRow(4)

	assert.NoError(t, c.Render(load(t, "sc"), "", &r))
	assert.Equal(t, "-/standard", r.ID)
	assert.Equal(t, render.Fields{"standard (default)", "kubernetes.io/gce-pd"}, r.Fields[:2])
}

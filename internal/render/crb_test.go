package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestClusterRoleBindingRender(t *testing.T) {
	c := render.ClusterRoleBinding{}
	r := render.NewRow(5)
	c.Render(load(t, "crb"), "-", &r)

	assert.Equal(t, "blee", r.ID)
	assert.Equal(t, render.Fields{"blee", "blee", "USR", "fernand"}, r.Fields[:4])
}

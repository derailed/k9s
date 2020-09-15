package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPolicyRender(t *testing.T) {
	var p render.Policy

	var r render.Row
	o := render.PolicyRes{
		Namespace:      "blee",
		Binding:        "fred",
		Resource:       "res",
		Group:          "grp",
		ResourceName:   "bob",
		NonResourceURL: "/blee",
		Verbs:          []string{"get", "list", "watch"},
	}

	assert.Nil(t, p.Render(o, "fred", &r))
	assert.Equal(t, "blee/res", r.ID)
	assert.Equal(t, render.Fields{
		"blee",
		"res",
		"grp",
		"fred",
		"[green::b] ✓ [::]",
		"[green::b] ✓ [::]",
		"[green::b] ✓ [::]",
		"[orangered::b] × [::]",
		"[orangered::b] × [::]",
		"[orangered::b] × [::]",
		"[orangered::b] × [::]",
		"[orangered::b] × [::]",
		"",
		"",
	}, r.Fields)
}

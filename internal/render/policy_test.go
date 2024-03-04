// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPolicyResMerge(t *testing.T) {
	uu := map[string]struct {
		p1, p2, e render.PolicyRes
		err       error
	}{
		"simple": {
			p1: render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get"}),
			p2: render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"patch"}),
			e:  render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get", "patch"}),
		},
		"dups": {
			p1: render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get"}),
			p2: render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get", "delete"}),
			e:  render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get", "delete"}),
		},
		"mismatch": {
			p1:  render.NewPolicyRes("fred", "blee", "deployments", "apps/v1", []string{"get"}),
			p2:  render.NewPolicyRes("fred", "blee", "statefulsets", "apps/v1", []string{"get", "delete"}),
			err: errors.New("policy mismatch apps/v1/deployments vs apps/v1/statefulsets"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			e, err := u.p1.Merge(u.p2)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, e)
		})
	}
}

func TestPolicyRender(t *testing.T) {
	var p render.Policy

	var r model1.Row
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
	assert.Equal(t, model1.Fields{
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

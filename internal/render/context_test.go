// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestContextHeader(t *testing.T) {
	var c render.Context

	assert.Equal(t, 4, len(c.Header("")))
}

func TestContextRender(t *testing.T) {
	uu := map[string]struct {
		ctx *render.NamedContext
		e   model1.Row
	}{
		"active": {
			ctx: &render.NamedContext{
				Name: "c1",
				Context: &api.Context{
					LocationOfOrigin: "fred",
					Cluster:          "c1",
					AuthInfo:         "u1",
					Namespace:        "ns1",
				},
				Config: &config{},
			},
			e: model1.Row{
				ID:     "c1",
				Fields: model1.Fields{"c1", "c1", "u1", "ns1"},
			},
		},
	}

	var r render.Context
	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			row := model1.NewRow(4)
			err := r.Render(uc.ctx, "", &row)

			assert.Nil(t, err)
			assert.Equal(t, uc.e, row)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type config struct{}

func (k config) CurrentContextName() (string, error) {
	return "fred", nil
}

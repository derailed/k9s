package render_test

import (
	"testing"

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
		ctx *api.Context
		e   render.Row
	}{
		"active": {
			ctx: &api.Context{
				LocationOfOrigin: "fred",
				Cluster:          "c1",
				AuthInfo:         "u1",
				Namespace:        "ns1",
			},
			e: render.Row{
				Fields: render.Fields{"", "c1", "u1", "ns1"},
			},
		},
	}

	var r render.Context
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			row := render.NewRow(4)
			err := r.Render(u.ctx, "", &row)

			assert.Nil(t, err)
			assert.Equal(t, u.e, row)
		})
	}
}

// Helpers...

func newContext(n string) *api.Context {
	return &api.Context{
		Cluster:   n,
		AuthInfo:  "blee",
		Namespace: "zorg",
	}
}

type config struct{}

func (k config) CurrentContextName() (string, error) {
	return "fred", nil
}

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPortForwardRender(t *testing.T) {
	var p render.PortForward
	var r render.Row
	o := render.ForwardRes{
		Forwarder: fwd{},
		Config: render.BenchCfg{
			C:    1,
			N:    1,
			Host: "0.0.0.0",
			Path: "/",
		},
	}

	assert.Nil(t, p.Render(o, "fred", &r))
	assert.Equal(t, "blee/fred", r.ID)
	assert.Equal(t, render.Fields{
		"blee",
		"fred",
		"co",
		"p1:p2",
		"http://0.0.0.0:p1/",
		"1",
		"1",
		"",
		"2m",
	}, r.Fields)
}

// Helpers...

type fwd struct{}

func (f fwd) ID() string {
	return "blee/fred"
}

func (f fwd) Path() string {
	return "blee/fred"
}

func (f fwd) Container() string {
	return "co"
}

func (f fwd) Port() string {
	return "p1:p2"
}

func (f fwd) Active() bool {
	return true
}

func (f fwd) Age() string {
	return "2m"
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortForwardRender(t *testing.T) {
	o := render.ForwardRes{
		Forwarder: fwd{},
		Config: render.BenchCfg{
			C:    1,
			N:    1,
			Host: "0.0.0.0",
			Path: "/",
		},
	}

	var p render.PortForward
	var r model1.Row
	require.NoError(t, p.Render(o, "fred", &r))
	assert.Equal(t, "blee/fred", r.ID)
	assert.Equal(t, model1.Fields{
		"blee",
		"fred",
		"co",
		"p1:p2",
		"http://0.0.0.0:p1/",
		"1",
		"1",
		"",
	}, r.Fields[:8])
}

// Helpers...

type fwd struct{}

func (fwd) ID() string {
	return "blee/fred"
}

func (fwd) Path() string {
	return "blee/fred"
}

func (fwd) Container() string {
	return "co"
}

func (fwd) Port() string {
	return "p1:p2"
}

func (fwd) Active() bool {
	return true
}

func (fwd) Age() time.Time {
	return testTime()
}

func (fwd) Address() string {
	return ""
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDpRender(t *testing.T) {
	c := render.Deployment{}
	r := model1.NewRow(7)

	require.NoError(t, c.Render(load(t, "dp"), "", &r))
	assert.Equal(t, "icx/icx-db", r.ID)
	assert.Equal(t, model1.Fields{"icx", "icx-db", "n/a", "1/1", "1", "1"}, r.Fields[:6])
}

func BenchmarkDpRender(b *testing.B) {
	var (
		c = render.Deployment{}
		r = model1.NewRow(7)
		o = load(b, "dp")
	)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = c.Render(o, "", &r)
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDpRender(t *testing.T) {
	c := render.Deployment{}
	r := render.NewRow(7)

	assert.Nil(t, c.Render(load(t, "dp"), "", &r))
	assert.Equal(t, "icx/icx-db", r.ID)
	assert.Equal(t, render.Fields{"icx", "icx-db", "0", "1/1", "1", "1"}, r.Fields[:6])
}

func BenchmarkDpRender(b *testing.B) {
	c := render.Deployment{}
	r := render.NewRow(7)
	o := load(b, "dp")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = c.Render(o, "", &r)
	}
}

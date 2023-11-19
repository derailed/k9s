// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestStorageClassRender(t *testing.T) {
	c := render.StorageClass{}
	r := render.NewRow(4)

	assert.NoError(t, c.Render(load(t, "sc"), "", &r))
	assert.Equal(t, "-/standard", r.ID)
	assert.Equal(t, render.Fields{"standard (default)", "kubernetes.io/gce-pd", "Delete", "Immediate", "true"}, r.Fields[:5])
}

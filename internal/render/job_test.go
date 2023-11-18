// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestJobRender(t *testing.T) {
	c := render.Job{}
	r := render.NewRow(4)

	assert.NoError(t, c.Render(load(t, "job"), "", &r))
	assert.Equal(t, "default/hello-1567179180", r.ID)
	assert.Equal(t, render.Fields{"default", "hello-1567179180", "1/1", "8s", "controller-uid=7473e6d0-cb3b-11e9-990f-42010a800218", "c1", "blang/busybox-bash"}, r.Fields[:7])
}

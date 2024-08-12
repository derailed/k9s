// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDaemonSetRender(t *testing.T) {
	c := render.DaemonSet{}
	r := model1.NewRow(9)

	assert.NoError(t, c.Render(load(t, "ds"), "", &r))
	assert.Equal(t, "kube-system/fluentd-gcp-v3.2.0", r.ID)
	assert.Equal(t, model1.Fields{"kube-system", "fluentd-gcp-v3.2.0", "0", "2", "2", "2", "2", "2"}, r.Fields[:8])
}

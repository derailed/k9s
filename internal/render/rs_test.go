// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestReplicaSetRender(t *testing.T) {
	c := render.ReplicaSet{}
	r := model1.NewRow(4)

	assert.NoError(t, c.Render(load(t, "rs"), "", &r))
	assert.Equal(t, "icx/icx-db-7d4b578979", r.ID)
	assert.Equal(t, model1.Fields{"icx", "icx-db-7d4b578979", "0", "1", "1", "1"}, r.Fields[:6])
}

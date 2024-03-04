// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestNetworkPolicyRender(t *testing.T) {
	c := render.NetworkPolicy{}
	r := model1.NewRow(9)

	assert.NoError(t, c.Render(load(t, "np"), "", &r))
	assert.Equal(t, "default/fred", r.ID)
	assert.Equal(t, model1.Fields{"default", "fred", "ns:app=blee,po:app=fred", "TCP:6379", "172.17.0.0/16[172.17.1.0/24,172.17.3.0/24...]", "", "TCP:5978", "10.0.0.0/24"}, r.Fields[:8])
}

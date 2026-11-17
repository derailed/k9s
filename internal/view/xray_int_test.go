// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestXrayGotoCmdBeforeTreeLoaded(t *testing.T) {
	v, ok := NewXray(client.NewGVR("deployments")).(*Xray)
	assert.True(t, ok)
	v.CmdBuff().SetActive(true)

	assert.NotPanics(t, func() {
		v.gotoCmd(nil)
	})
	assert.False(t, v.CmdBuff().IsActive())
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestConfigMapNew(t *testing.T) {
	s := view.NewConfigMap(client.NewGVR("v1/configmaps"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "ConfigMaps", s.Name())
	assert.Equal(t, 7, len(s.Hints()))
}

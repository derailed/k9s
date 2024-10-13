// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestContainerNew(t *testing.T) {
	c := view.NewContainer(client.NewGVR("containers"))

	assert.Nil(t, c.Init(makeCtx()))
	assert.Equal(t, "Containers", c.Name())
	assert.Equal(t, 19, len(c.Hints()))
}

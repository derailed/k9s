// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPriorityClassNew(t *testing.T) {
	s := view.NewPriorityClass(client.NewGVR("scheduling.k8s.io/v1/priorityclasses"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "PriorityClass", s.Name())
	assert.Equal(t, 6, len(s.Hints()))
}

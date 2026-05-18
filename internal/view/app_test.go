// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	a := view.NewApp(mock.NewMockConfig(t))
	_ = a.Init("blee", 10)

	assert.Equal(t, 15, a.GetActions().Len())
	_, ok := a.GetActions().Get(ui.KeyShiftT)
	assert.True(t, ok)
}

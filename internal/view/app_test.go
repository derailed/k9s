// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	a := view.NewApp(mock.NewMockConfig())
	_ = a.Init("blee", 10)

	assert.Equal(t, 12, a.GetActions().Len())
}

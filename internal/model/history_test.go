// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	h := model.NewHistory(3)
	for i := 1; i < 5; i++ {
		h.Push(fmt.Sprintf("cmd%d", i))
	}

	assert.Equal(t, []string{"cmd4", "cmd3", "cmd2"}, h.List())
	h.Clear()
	assert.True(t, h.Empty())
}

func TestHistoryDups(t *testing.T) {
	h := model.NewHistory(3)
	for i := 1; i < 4; i++ {
		h.Push(fmt.Sprintf("cmd%d", i))
	}
	h.Push("cmd1")
	h.Push("")

	assert.Equal(t, []string{"cmd3", "cmd2", "cmd1"}, h.List())
}

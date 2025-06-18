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
		h.Push(fmt.Sprintf("cmd%d", i), fmt.Sprintf("filter%d", i))
	}

	assert.Equal(t, []model.FilteredCommand{
		{
			Command: "cmd1",
			Filter:  "filter1",
		},
		{
			Command: "cmd2",
			Filter:  "filter2",
		},
		{
			Command: "cmd3",
			Filter:  "filter3",
		},
	}, h.List())
	h.Clear()
	assert.True(t, h.Empty())
}

func TestHistoryDups(t *testing.T) {
	h := model.NewHistory(3)
	for i := 1; i < 4; i++ {
		h.Push(fmt.Sprintf("cmd%d", i), fmt.Sprintf("filter%d", i))
	}
	h.Push("cmd1", "filter1")
	h.Push("cmd1", "")
	h.Push("cmd1", "")
	h.Push("", "")

	assert.Equal(t, []model.FilteredCommand{
		{
			Command: "cmd1",
			Filter:  "filter1",
		},
		{
			Command: "cmd2",
			Filter:  "filter2",
		},
		{
			Command: "cmd3",
			Filter:  "filter3",
		},
	}, h.List())
}

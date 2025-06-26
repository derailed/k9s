// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestHistoryClear(t *testing.T) {
	h := model.NewHistory(3)
	for i := 1; i < 5; i++ {
		h.Push(fmt.Sprintf("cmd%d", i))
	}
	assert.Equal(t, []string{"cmd1", "cmd2", "cmd3"}, h.List())

	h.Clear()
	assert.True(t, h.Empty())
}

func TestHistoryPush(t *testing.T) {
	h := model.NewHistory(3)
	for i := 1; i < 4; i++ {
		h.Push(fmt.Sprintf("cmd%d", i))
	}
	h.Push("cmd1")
	h.Push("")

	assert.Equal(t, []string{"cmd1", "cmd2", "cmd3"}, h.List())
}

func TestHistoryTop(t *testing.T) {
	uu := map[string]struct {
		push []string
		pop  int
		cmd  string
		ok   bool
	}{
		"empty": {},

		"no-one-left": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  3,
		},

		"last": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			cmd:  "cmd3",
			ok:   true,
		},

		"middle": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  1,
			cmd:  "cmd2",
			ok:   true,
		},

		"first": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  2,
			cmd:  "cmd1",
			ok:   true,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			h := model.NewHistory(3)
			for _, cmd := range u.push {
				h.Push(cmd)
			}
			for range u.pop {
				_ = h.Pop()
			}

			cmd, ok := h.Top()
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.cmd, cmd)
		})
	}
}

func TestHistoryBack(t *testing.T) {
	uu := map[string]struct {
		push []string
		pop  int
		cmd  string
		ok   bool
	}{
		"empty": {},

		"pop-all": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  3,
		},

		"pop-none": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			cmd:  "cmd2",
			ok:   true,
		},

		"pop-one": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  1,
			cmd:  "cmd1",
			ok:   true,
		},

		"pop-to-first": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			pop:  2,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			h := model.NewHistory(3)
			for _, cmd := range u.push {
				h.Push(cmd)
			}
			for range u.pop {
				_ = h.Pop()
			}

			cmd, ok := h.Back()
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.cmd, cmd)
		})
	}
}

func TestHistoryForward(t *testing.T) {
	uu := map[string]struct {
		push []string
		back int
		cmd  string
		ok   bool
	}{
		"empty": {},

		"back-2": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			back: 2,
			cmd:  "cmd2",
			ok:   true,
		},

		"back-1": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			back: 1,
			cmd:  "cmd3",
			ok:   true,
		},

		"back-all": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			back: 3,
			cmd:  "cmd2",
			ok:   true,
		},

		"back-none": {
			push: []string{"cmd1", "cmd2", "cmd3"},
			back: 0,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			h := model.NewHistory(3)
			for _, cmd := range u.push {
				h.Push(cmd)
			}
			for range u.back {
				_, _ = h.Back()
			}

			cmd, ok := h.Forward()
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.cmd, cmd)
		})
	}
}

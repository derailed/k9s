// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestPagesPush(t *testing.T) {
	c1, c2 := makeComponent("c1"), makeComponent("c2")

	p := ui.NewPages()
	p.Push(c1)
	p.Push(c2)

	assert.Equal(t, 2, p.GetPageCount())
	assert.Equal(t, c2, p.CurrentPage().Item)
}

func TestPagesPop(t *testing.T) {
	c1, c2 := makeComponent("c1"), makeComponent("c2")

	p := ui.NewPages()
	p.Push(c1)
	p.Push(c2)
	p.Pop()

	assert.Equal(t, 1, p.GetPageCount())
	assert.Equal(t, c1, p.CurrentPage().Item)
}

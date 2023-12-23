// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestNewCrumbs(t *testing.T) {
	v := ui.NewCrumbs(config.NewStyles())
	v.StackPushed(makeComponent("c1"))
	v.StackPushed(makeComponent("c2"))
	v.StackPushed(makeComponent("c3"))

	assert.Equal(t, "[#000000:#00ffff:b] <c1> [-:#000000:-] [#000000:#00ffff:b] <c2> [-:#000000:-] [#000000:#ffa500:b] <c3> [-:#000000:-] \n", v.GetText(false))
}

// Helpers...

type c struct {
	name string
}

func makeComponent(n string) c {
	return c{name: n}
}

func (c) InCmdMode() bool                                              { return false }
func (c c) HasFocus() bool                                             { return true }
func (c c) Hints() model.MenuHints                                     { return nil }
func (c c) ExtraHints() map[string]string                              { return nil }
func (c c) Name() string                                               { return c.name }
func (c c) Draw(tcell.Screen)                                          {}
func (c c) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) { return nil }
func (c c) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return nil
}
func (c c) SetRect(int, int, int, int)       {}
func (c c) GetRect() (int, int, int, int)    { return 0, 0, 0, 0 }
func (c c) GetFocusable() tview.Focusable    { return c }
func (c c) Focus(func(tview.Primitive))      {}
func (c c) Blur()                            {}
func (c c) Start()                           {}
func (c c) Stop()                            {}
func (c c) Init(context.Context) error       { return nil }
func (c c) SetFilter(string)                 {}
func (c c) SetLabelFilter(map[string]string) {}

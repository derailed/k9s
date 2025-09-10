// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
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

func (c) SetCommand(*cmd.Interpreter)                                {}
func (c) InCmdMode() bool                                            { return false }
func (c) HasFocus() bool                                             { return true }
func (c) Hints() model.MenuHints                                     { return nil }
func (c) ExtraHints() map[string]string                              { return nil }
func (c c) Name() string                                             { return c.name }
func (c) Draw(tcell.Screen)                                          {}
func (c) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) { return nil }
func (c) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return nil
}
func (c) SetRect(int, int, int, int)       {}
func (c) GetRect() (a, b, c, d int)        { return 0, 0, 0, 0 }
func (c c) GetFocusable() tview.Focusable  { return c }
func (c) Focus(func(tview.Primitive))      {}
func (c) Blur()                            {}
func (c) Start()                           {}
func (c) Stop()                            {}
func (c) Init(context.Context) error       { return nil }
func (c) SetFilter(string)                 {}
func (c) SetLabelSelector(labels.Selector) {}

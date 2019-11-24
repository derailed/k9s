package ui_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestNewCrumbs(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := ui.NewCrumbs(defaults)
	v.StackPushed(makeComponent("c1"))
	v.StackPushed(makeComponent("c2"))
	v.StackPushed(makeComponent("c3"))

	assert.Equal(t, "[black:aqua:b] <c1> [-:black:-] [black:aqua:b] <c2> [-:black:-] [black:orange:b] <c3> [-:black:-] \n", v.GetText(false))
}

// Helpers...

type c struct {
	name string
}

func makeComponent(n string) c {
	return c{name: n}
}

func (c c) HasFocus() bool                                             { return true }
func (c c) Hints() model.MenuHints                                     { return nil }
func (c c) Name() string                                               { return c.name }
func (c c) Draw(tcell.Screen)                                          {}
func (c c) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) { return nil }
func (c c) SetRect(int, int, int, int)                                 {}
func (c c) GetRect() (int, int, int, int)                              { return 0, 0, 0, 0 }
func (c c) GetFocusable() tview.Focusable                              { return c }
func (c c) Focus(func(tview.Primitive))                                {}
func (c c) Blur()                                                      {}
func (c c) Start()                                                     {}
func (c c) Stop()                                                      {}
func (c c) Init(context.Context) error                                 { return nil }

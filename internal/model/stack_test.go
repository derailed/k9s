package model_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestStackPush(t *testing.T) {
	top := c{}
	uu := map[string]struct {
		items []model.Component
		pop   int
		e     bool
		top   model.Component
	}{
		"empty": {
			items: []model.Component{},
			pop:   3,
			e:     true,
		},
		"full": {
			items: []model.Component{c{}, c{}, top},
			pop:   3,
			e:     true,
		},
		"pop": {
			items: []model.Component{c{}, c{}, top},
			pop:   2,
			e:     false,
			top:   top,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			for i := 0; i < u.pop; i++ {
				s.Pop()
			}
			assert.Equal(t, u.e, s.Empty())
			if !u.e {
				assert.Equal(t, u.top, s.Top())
			}
		})
	}
}

func TestStackTop(t *testing.T) {
	top := c{}
	uu := map[string]struct {
		items []model.Component
		e     model.Component
	}{
		"blank": {
			items: []model.Component{},
		},
		"push3": {
			items: []model.Component{c{}, c{}, top},
			e:     top,
		},
		"push1": {
			items: []model.Component{top},
			e:     top,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, item := range u.items {
				s.Push(item)
			}
			v := s.Top()
			assert.Equal(t, u.e, v)
		})
	}
}

func TestStackListener(t *testing.T) {
	items := []model.Component{c{}, c{}, c{}}
	s := model.NewStack()
	l := stackL{}
	s.AddListener(&l)
	for _, item := range items {
		s.Push(item)
	}
	assert.Equal(t, 3, l.count)

	for range items {
		s.Pop()
	}
	assert.Equal(t, 0, l.count)
}

func TestStackRemoveListener(t *testing.T) {
	s := model.NewStack()
	l1, l2, l3 := &stackL{}, &stackL{}, &stackL{}
	s.AddListener(l1)
	s.AddListener(l2)
	s.AddListener(l3)

	s.RemoveListener(l2)
	s.RemoveListener(l3)
	s.RemoveListener(l1)

	s.Push(c{})

	assert.Equal(t, 0, l1.count)
	assert.Equal(t, 0, l2.count)
	assert.Equal(t, 0, l3.count)
}

type stackL struct {
	count int
}

func (s *stackL) StackPushed(model.Component) {
	s.count++
}
func (s *stackL) StackPopped(c, top model.Component) {
	s.count--
}
func (s *stackL) StackTop(model.Component) {}

type c struct{}

func (c c) Name() string                                               { return "test" }
func (c c) Hints() model.MenuHints                                     { return nil }
func (c c) Draw(tcell.Screen)                                          {}
func (c c) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) { return nil }
func (c c) SetRect(int, int, int, int)                                 {}
func (c c) GetRect() (int, int, int, int)                              { return 0, 0, 0, 0 }
func (c c) GetFocusable() tview.Focusable                              { return nil }
func (c c) Focus(func(tview.Primitive))                                {}
func (c c) Blur()                                                      {}
func (c c) Start()                                                     {}
func (c c) Stop()                                                      {}
func (c c) Init(context.Context)                                       {}

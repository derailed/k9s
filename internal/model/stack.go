// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	// StackPush denotes an add on the stack.
	StackPush StackAction = 1 << iota

	// StackPop denotes a delete on the stack.
	StackPop
)

// StackAction represents an action on the stack.
type StackAction int

// StackEvent represents an operation on a view stack.
type StackEvent struct {
	// Kind represents the event condition.
	Action StackAction

	// Item represents the targeted item.
	Component Component
}

// StackListener represents a stack listener.
type StackListener interface {
	// StackPushed indicates a new item was added.
	StackPushed(Component)

	// StackPopped indicates an item was deleted
	StackPopped(old, new Component)

	// StackTop indicates the top of the stack
	StackTop(Component)
}

// Stack represents a stacks of components.
type Stack struct {
	components []Component
	listeners  []StackListener
	mx         sync.RWMutex
}

// NewStack returns a new initialized stack.
func NewStack() *Stack {
	return &Stack{}
}

// Flatten returns a string representation of the component stack.
func (s *Stack) Flatten() []string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	ss := make([]string, len(s.components))
	for i, c := range s.components {
		ss[i] = c.Name()
	}
	return ss
}

// RemoveListener removes a listener.
func (s *Stack) RemoveListener(l StackListener) {
	victim := -1
	for i, lis := range s.listeners {
		if lis == l {
			victim = i
			break
		}
	}
	if victim == -1 {
		return
	}
	s.listeners = append(s.listeners[:victim], s.listeners[victim+1:]...)
}

// AddListener registers a stack listener.
func (s *Stack) AddListener(l StackListener) {
	s.listeners = append(s.listeners, l)
	if !s.Empty() {
		l.StackTop(s.Top())
	}
}

// Push adds a new item.
func (s *Stack) Push(c Component) {
	if top := s.Top(); top != nil {
		top.Stop()
	}

	s.mx.Lock()
	{
		s.components = append(s.components, c)
	}
	s.mx.Unlock()
	s.notify(StackPush, c)
}

// Pop removed the top item and returns it.
func (s *Stack) Pop() (Component, bool) {
	if s.Empty() {
		return nil, false
	}

	var c Component
	s.mx.Lock()
	{
		c = s.components[len(s.components)-1]
		c.Stop()
		s.components = s.components[:len(s.components)-1]
	}
	s.mx.Unlock()
	s.notify(StackPop, c)

	return c, true
}

// Peek returns stack state.
func (s *Stack) Peek() []Component {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.components
}

// Clear clear out the stack using pops.
func (s *Stack) Clear() {
	for range s.components {
		s.Pop()
	}
}

// Empty returns true if the stack is empty.
func (s *Stack) Empty() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return len(s.components) == 0
}

// IsLast indicates if stack only has one item left.
func (s *Stack) IsLast() bool {
	return len(s.components) == 1
}

// Previous returns the previous component if any.
func (s *Stack) Previous() Component {
	if s.Empty() {
		return nil
	}
	if s.IsLast() {
		return s.Top()
	}

	return s.components[len(s.components)-2]
}

// Top returns the top most item or nil if the stack is empty.
func (s *Stack) Top() Component {
	if s.Empty() {
		return nil
	}

	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.components[len(s.components)-1]
}

func (s *Stack) notify(a StackAction, c Component) {
	for _, l := range s.listeners {
		switch a {
		case StackPush:
			l.StackPushed(c)
		case StackPop:
			l.StackPopped(c, s.Top())
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

// Dump prints out the stack.
func (s *Stack) Dump() {
	log.Debug().Msgf("--- Stack Dump %p---", s)
	for i, c := range s.components {
		log.Debug().Msgf("%d -- %s -- %#v", i, c.Name(), c)
	}
	log.Debug().Msg("------------------")
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"sync"
	"time"
)

const (
	maxBuff = 10

	keyEntryDelay = 100 * time.Millisecond

	// CommandBuffer represents a command buffer.
	CommandBuffer BufferKind = 1 << iota
	// FilterBuffer represents a filter buffer.
	FilterBuffer
)

type (
	// BufferKind indicates a buffer type.
	BufferKind int8

	// BuffWatcher represents a command buffer listener.
	BuffWatcher interface {
		// BufferCompleted indicates input was accepted.
		BufferCompleted(text, suggestion string)

		// BufferChanged indicates the buffer was changed.
		BufferChanged(text, suggestion string)

		// BufferActive indicates the buff activity changed.
		BufferActive(state bool, kind BufferKind)
	}
)

// CmdBuff represents user command input.
type CmdBuff struct {
	buff       []rune
	suggestion string
	listeners  map[BuffWatcher]struct{}
	hotKey     rune
	kind       BufferKind
	active     bool
	cancel     context.CancelFunc
	mx         sync.RWMutex
}

// NewCmdBuff returns a new command buffer.
func NewCmdBuff(key rune, kind BufferKind) *CmdBuff {
	return &CmdBuff{
		hotKey:    key,
		kind:      kind,
		buff:      make([]rune, 0, maxBuff),
		listeners: make(map[BuffWatcher]struct{}),
	}
}

// InCmdMode checks if a command exists and the buffer is active.
func (c *CmdBuff) InCmdMode() bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	if !c.active {
		return false
	}

	return len(c.buff) > 0
}

// IsActive checks if command buffer is active.
func (c *CmdBuff) IsActive() bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.active
}

// SetActive toggles cmd buffer active state.
func (c *CmdBuff) SetActive(b bool) {
	c.mx.Lock()
	{
		c.active = b
	}
	c.mx.Unlock()

	c.fireActive(c.active)
}

// GetText returns the current text.
func (c *CmdBuff) GetText() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return string(c.buff)
}

// GetKind returns the buffer kind.
func (c *CmdBuff) GetKind() BufferKind {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.kind
}

// GetSuggestion returns the current suggestion.
func (c *CmdBuff) GetSuggestion() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.suggestion
}

func (c *CmdBuff) hasCancel() bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.cancel != nil
}

func (c *CmdBuff) setCancel(f context.CancelFunc) {
	c.mx.Lock()
	{
		c.cancel = f
	}
	c.mx.Unlock()
}

func (c *CmdBuff) resetCancel() {
	c.mx.Lock()
	{
		c.cancel = nil
	}
	c.mx.Unlock()
}

// SetText initializes the buffer with a command.
func (c *CmdBuff) SetText(text, suggestion string) {
	c.mx.Lock()
	{
		c.buff, c.suggestion = []rune(text), suggestion
	}
	c.mx.Unlock()
	c.fireBufferCompleted(c.GetText(), c.GetSuggestion())
}

// Add adds a new character to the buffer.
func (c *CmdBuff) Add(r rune) {
	c.mx.Lock()
	{
		c.buff = append(c.buff, r)
	}
	c.mx.Unlock()
	c.fireBufferChanged(c.GetText(), c.GetSuggestion())
	if c.hasCancel() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), keyEntryDelay)
	c.setCancel(cancel)

	go func() {
		<-ctx.Done()
		c.fireBufferCompleted(c.GetText(), c.GetSuggestion())
		c.resetCancel()
	}()
}

// Delete removes the last character from the buffer.
func (c *CmdBuff) Delete() {
	if c.Empty() {
		return
	}
	c.SetText(string(c.buff[:len(c.buff)-1]), "")
	c.fireBufferChanged(c.GetText(), c.GetSuggestion())
	if c.hasCancel() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	c.setCancel(cancel)

	go func() {
		<-ctx.Done()
		c.fireBufferCompleted(c.GetText(), c.GetSuggestion())
		c.resetCancel()
	}()
}

// ClearText clears out command buffer.
func (c *CmdBuff) ClearText(fire bool) {
	c.mx.Lock()
	{
		c.buff, c.suggestion = c.buff[:0], ""
	}
	c.mx.Unlock()

	if fire {
		c.fireBufferCompleted(c.GetText(), c.GetSuggestion())
	}
}

// Reset clears out the command buffer and deactivates it.
func (c *CmdBuff) Reset() {
	c.ClearText(true)
	c.SetActive(false)
	c.fireBufferCompleted(c.GetText(), c.GetSuggestion())
}

// Empty returns true if no cmd, false otherwise.
func (c *CmdBuff) Empty() bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return len(c.buff) == 0
}

// ----------------------------------------------------------------------------
// Event Listeners...

// AddListener registers a cmd buffer listener.
func (c *CmdBuff) AddListener(w BuffWatcher) {
	c.mx.Lock()
	{
		c.listeners[w] = struct{}{}
	}
	c.mx.Unlock()
}

// RemoveListener removes a listener.
func (c *CmdBuff) RemoveListener(l BuffWatcher) {
	c.mx.Lock()
	delete(c.listeners, l)
	c.mx.Unlock()
}

func (c *CmdBuff) fireBufferCompleted(t, s string) {
	for l := range c.listeners {
		l.BufferCompleted(t, s)
	}
}

func (c *CmdBuff) fireBufferChanged(t, s string) {
	for l := range c.listeners {
		l.BufferChanged(t, s)
	}
}

func (c *CmdBuff) fireActive(b bool) {
	for l := range c.listeners {
		l.BufferActive(b, c.GetKind())
	}
}

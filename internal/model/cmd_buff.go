package model

import (
	"context"
	"sync"
	"time"
)

const (
	maxBuff = 10

	keyEntryDelay = 200 * time.Millisecond

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
		BufferCompleted(s string)

		// BufferChanged indicates the buffer was changed.
		BufferChanged(s string)

		// BufferActive indicates the buff activity changed.
		BufferActive(state bool, kind BufferKind)
	}
)

// CmdBuff represents user command input.
type CmdBuff struct {
	buff      []rune
	listeners []BuffWatcher
	hotKey    rune
	kind      BufferKind
	active    bool
	cancel    context.CancelFunc
	mx        sync.RWMutex
}

// NewCmdBuff returns a new command buffer.
func NewCmdBuff(key rune, kind BufferKind) *CmdBuff {
	return &CmdBuff{
		hotKey:    key,
		kind:      kind,
		buff:      make([]rune, 0, maxBuff),
		listeners: []BuffWatcher{},
	}
}

// InCmdMode checks if a command exists and the buffer is active.
func (c *CmdBuff) InCmdMode() bool {
	return c.active || len(c.buff) > 0
}

// IsActive checks if command buffer is active.
func (c *CmdBuff) IsActive() bool {
	return c.active
}

// SetActive toggles cmd buffer active state.
func (c *CmdBuff) SetActive(b bool) {
	c.active = b
	c.fireActive(c.active)
}

// GetText returns the current text.
func (c *CmdBuff) GetText() string {
	return string(c.buff)
}

// SetText initializes the buffer with a command.
func (c *CmdBuff) SetText(cmd string) {
	c.buff = []rune(cmd)
	c.fireBufferCompleted()
}

// Add adds a new character to the buffer.
func (c *CmdBuff) Add(r rune) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.buff = append(c.buff, r)
	c.fireBufferChanged()
	if c.cancel != nil {
		return
	}
	var ctx context.Context
	ctx, c.cancel = context.WithTimeout(context.Background(), keyEntryDelay)

	go func() {
		<-ctx.Done()
		c.mx.Lock()
		{
			c.fireBufferCompleted()
			c.cancel = nil
		}
		c.mx.Unlock()
	}()
}

// Delete removes the last character from the buffer.
func (c *CmdBuff) Delete() {
	c.mx.Lock()
	defer c.mx.Unlock()
	if c.Empty() {
		return
	}
	c.buff = c.buff[:len(c.buff)-1]
	c.fireBufferChanged()
	if c.cancel != nil {
		return
	}

	var ctx context.Context
	ctx, c.cancel = context.WithTimeout(context.Background(), 800*time.Millisecond)

	go func() {
		<-ctx.Done()
		c.mx.Lock()
		{
			c.fireBufferCompleted()
			c.cancel = nil
		}
		c.mx.Unlock()
	}()
}

// ClearText clears out command buffer.
func (c *CmdBuff) ClearText(fire bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.buff = make([]rune, 0, maxBuff)
	if fire {
		c.fireBufferCompleted()
	}
}

// Reset clears out the command buffer and deactivates it.
func (c *CmdBuff) Reset() {
	c.ClearText(true)
	c.SetActive(false)
	c.fireBufferCompleted()
}

// Empty returns true if no cmd, false otherwise.
func (c *CmdBuff) Empty() bool {
	return len(c.buff) == 0
}

// ----------------------------------------------------------------------------
// Event Listeners...

// AddListener registers a cmd buffer listener.
func (c *CmdBuff) AddListener(w BuffWatcher) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.listeners = append(c.listeners, w)
}

// RemoveListener removes a listener.
func (c *CmdBuff) RemoveListener(l BuffWatcher) {
	c.mx.Lock()
	defer c.mx.Unlock()

	victim := -1
	for i, lis := range c.listeners {
		if l == lis {
			victim = i
			break
		}
	}

	if victim == -1 {
		return
	}
	c.listeners = append(c.listeners[:victim], c.listeners[victim+1:]...)
}

func (c *CmdBuff) fireBufferCompleted() {
	text := c.GetText()
	for _, l := range c.listeners {
		l.BufferCompleted(text)
	}
}

func (c *CmdBuff) fireBufferChanged() {
	text := c.GetText()
	for _, l := range c.listeners {
		l.BufferChanged(text)
	}
}

func (c *CmdBuff) fireActive(b bool) {
	for _, l := range c.listeners {
		l.BufferActive(b, c.kind)
	}
}

package model

const maxBuff = 10

const (
	// CommandBuffer represents a command buffer.
	CommandBuffer BufferKind = 1 << iota
	// FilterBuffer represents a filter buffer.
	FilterBuffer
)

// BufferKind indicates a buffer type
type BufferKind int8

// BuffWatcher represents a command buffer listener.
type BuffWatcher interface {
	// Changed indicates the buffer was changed.
	BufferChanged(s string)

	// Active indicates the buff activity changed.
	BufferActive(state bool, kind BufferKind)
}

// CmdBuff represents user command input.
type CmdBuff struct {
	buff      []rune
	listeners []BuffWatcher
	hotKey    rune
	kind      BufferKind
	active    bool
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
	c.fireBufferChanged()
}

// Add adds a new charater to the buffer.
func (c *CmdBuff) Add(r rune) {
	c.buff = append(c.buff, r)
	c.fireBufferChanged()
}

// Delete removes the last character from the buffer.
func (c *CmdBuff) Delete() {
	if c.Empty() {
		return
	}
	c.buff = c.buff[:len(c.buff)-1]
	c.fireBufferChanged()
}

// ClearText clears out command buffer.
func (c *CmdBuff) ClearText(fire bool) {
	c.buff = make([]rune, 0, maxBuff)
	if fire {
		c.fireBufferChanged()
	}
}

// Reset clears out the command buffer and deactivates it.
func (c *CmdBuff) Reset() {
	c.ClearText(true)
	c.SetActive(false)
}

// Empty returns true if no cmd, false otherwise.
func (c *CmdBuff) Empty() bool {
	return len(c.buff) == 0
}

// ----------------------------------------------------------------------------
// Event Listeners...

// AddListener registers a cmd buffer listener.
func (c *CmdBuff) AddListener(w BuffWatcher) {
	c.listeners = append(c.listeners, w)
}

// RemoveListener removes a listener.
func (c *CmdBuff) RemoveListener(l BuffWatcher) {
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

func (c *CmdBuff) fireBufferChanged() {
	for _, l := range c.listeners {
		l.BufferChanged(c.GetText())
	}
}

func (c *CmdBuff) fireActive(b bool) {
	for _, l := range c.listeners {
		l.BufferActive(b, c.kind)
	}
}

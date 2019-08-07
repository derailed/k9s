package ui

const maxBuff = 10

type (
	buffWatcher interface {
		changed(s string)
		active(state bool)
	}

	// CmdBuff represents user command input.
	CmdBuff struct {
		buff      []rune
		hotKey    rune
		active    bool
		listeners []buffWatcher
	}
)

// NewCmdBuff returns a new command buffer.
func NewCmdBuff(key rune) *CmdBuff {
	return &CmdBuff{
		hotKey:    key,
		buff:      make([]rune, 0, maxBuff),
		listeners: []buffWatcher{},
	}
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

// String turns rune to string (Stringer protocol)
func (c *CmdBuff) String() string {
	return string(c.buff)
}

// Set initializes the buffer with a command.
func (c *CmdBuff) Set(rr []rune) {
	c.buff = rr
}

// Add adds a new charater to the buffer.
func (c *CmdBuff) Add(r rune) {
	c.buff = append(c.buff, r)
	c.fireChanged()
}

// Delete removes the last character from the buffer.
func (c *CmdBuff) Delete() {
	if c.Empty() {
		return
	}
	c.buff = c.buff[:len(c.buff)-1]
	c.fireChanged()
}

func (c *CmdBuff) wipe() {
	c.buff = make([]rune, 0, maxBuff)
}

// Clear clears out command buffer.
func (c *CmdBuff) Clear() {
	c.buff = make([]rune, 0, maxBuff)
	c.fireChanged()
}

// Reset clears out the command buffer.
func (c *CmdBuff) Reset() {
	c.Clear()
	c.fireChanged()
	c.SetActive(false)
}

// Empty returns true is no cmd, false otherwise.
func (c *CmdBuff) Empty() bool {
	return len(c.buff) == 0
}

// ----------------------------------------------------------------------------
// Event Listeners...

// AddListener registers a cmd buffer listener.
func (c *CmdBuff) AddListener(w ...buffWatcher) {
	c.listeners = append(c.listeners, w...)
}

func (c *CmdBuff) fireChanged() {
	for _, l := range c.listeners {
		l.changed(c.String())
	}
}

func (c *CmdBuff) fireActive(b bool) {
	for _, l := range c.listeners {
		l.active(b)
	}
}

package views

import(
	log "github.com/sirupsen/logrus"
)

const maxBuff = 10

type buffWatcher interface {
	changed(s string)
	active(state bool)
}

type cmdBuff struct {
	buff      []rune
	hotKey    rune
	active    bool
	listeners []buffWatcher
}

func newCmdBuff(key rune) *cmdBuff {
	return &cmdBuff{
		hotKey:    key,
		buff:      make([]rune, 0, maxBuff),
		listeners: []buffWatcher{},
	}
}

func (c *cmdBuff) isActive() bool {
	log.Debugf("Cmd buff `%s Active:%t", string(c.hotKey), c.active)
	return c.active
}

func (c *cmdBuff) setActive(b bool) {
	log.Debugf("Cmd buff `%s SetActive:%t", string(c.hotKey), b)
	c.active = b
	c.fireActive(c.active)
}

// String turns rune to string (Stringer protocol)
func (c *cmdBuff) String() string {
	return string(c.buff)
}

func (c *cmdBuff) add(r rune) {
	c.buff = append(c.buff, r)
	c.fireChanged()
}

func (c *cmdBuff) del() {
	if c.empty() {
		return
	}
	c.buff = c.buff[:len(c.buff)-1]
	c.fireChanged()
}

func (c *cmdBuff) clear() {
	c.buff = make([]rune, 0, maxBuff)
	c.fireChanged()
}

func (c *cmdBuff) reset() {
	c.clear()
	c.fireChanged()
	c.setActive(false)
}

func (c *cmdBuff) empty() bool {
	return len(c.buff) == 0
}

// ----------------------------------------------------------------------------
// Event Listeners...

func (c *cmdBuff) addListener(w ...buffWatcher) {
	c.listeners = append(c.listeners, w...)
}

func (c *cmdBuff) fireChanged() {
	for _, l := range c.listeners {
		l.changed(c.String())
	}
}

func (c *cmdBuff) fireActive(b bool) {
	for _, l := range c.listeners {
		l.active(b)
	}
}

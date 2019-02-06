package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testListener struct {
	text  string
	act   int
	inact int
}

func (l *testListener) changed(s string) {
	l.text = s
}

func (l *testListener) active(s bool) {
	if s {
		l.act++
		return
	}
	l.inact++
}

func TestCmdBuffActivate(t *testing.T) {
	b, l := newCmdBuff('>'), testListener{}
	b.addListener(&l)

	b.setActive(true)
	assert.Equal(t, 1, l.act)
	assert.Equal(t, 0, l.inact)
	assert.True(t, b.active)
}

func TestCmdBuffDeactivate(t *testing.T) {
	b, l := newCmdBuff('>'), testListener{}
	b.addListener(&l)

	b.setActive(false)
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 1, l.inact)
	assert.False(t, b.active)
}

func TestCmdBuffChanged(t *testing.T) {
	b, l := newCmdBuff('>'), testListener{}
	b.addListener(&l)

	b.add('b')
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "b", l.text)
	assert.Equal(t, "b", b.String())

	b.del()
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.String())

	b.add('c')
	b.clear()
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.String())

	b.add('c')
	b.reset()
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 1, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.String())
	assert.True(t, b.empty())
}

func TestCmdBuffAdd(t *testing.T) {
	b := newCmdBuff('>')

	uu := []struct {
		runes []rune
		cmd   string
	}{
		{[]rune{}, ""},
		{[]rune{'a'}, "a"},
		{[]rune{'a', 'b', 'c'}, "abc"},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.add(r)
		}
		assert.Equal(t, u.cmd, b.String())
		b.reset()
	}
}

func TestCmdBuffDel(t *testing.T) {
	b := newCmdBuff('>')

	uu := []struct {
		runes []rune
		cmd   string
	}{
		{[]rune{}, ""},
		{[]rune{'a'}, ""},
		{[]rune{'a', 'b', 'c'}, "ab"},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.add(r)
		}
		b.del()
		assert.Equal(t, u.cmd, b.String())
		b.reset()
	}
}

func TestCmdBuffEmpty(t *testing.T) {
	b := newCmdBuff('>')

	uu := []struct {
		runes []rune
		empty bool
	}{
		{[]rune{}, true},
		{[]rune{'a'}, false},
		{[]rune{'a', 'b', 'c'}, false},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.add(r)
		}
		assert.Equal(t, u.empty, b.empty())
		b.reset()
	}
}

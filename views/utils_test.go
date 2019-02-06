package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogBufferAdd(t *testing.T) {
	uu := []struct {
		lines    []string
		expected []string
	}{
		{[]string{}, []string{}},
		{[]string{"l1"}, []string{"l1"}},
		{[]string{"l1", "l2"}, []string{"l1", "l2"}},
		{[]string{"l1", "l2", "l3"}, []string{"l2", "l3"}},
		{[]string{"l1", "l2", "l3", "l4"}, []string{"l3", "l4"}},
	}

	for _, u := range uu {
		b := newLogBuffer(2, false)
		for _, l := range u.lines {
			b.add(l)
		}

		assert.Equal(t, len(u.expected), b.length())
		assert.Equal(t, u.expected, b.lines())
	}
}

func TestLogBufferCleanse(t *testing.T) {
	b := newLogBuffer(2, true)
	ll := []string{"l1", "l2"}
	ee := []string{b.decorateLine("l1"), b.decorateLine("l2")}
	for _, l := range ll {
		b.add(l)
	}
	assert.Equal(t, ee, b.lines())
	b.cleanse()
	assert.Equal(t, ll, b.lines())
}

func TestLogBufferDecorate(t *testing.T) {
	l := "hello k9s"
	var b *logBuffer
	assert.Equal(t, "["+newLogColor+"::]"+l+"[::]", b.decorateLine(l))
}

func TestLogBufferTrimLine(t *testing.T) {
	l := "hello k9s"
	dl := "[" + newLogColor + "::]" + l + "[::]"
	b := newLogBuffer(1, true)
	assert.Equal(t, l, b.trimLine(dl))
}

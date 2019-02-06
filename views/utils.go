package views

import (
	"fmt"
	"regexp"
	"strings"
)

const newLogColor = "greenyellow"

type (
	logBuffer struct {
		capacity int
		decorate bool
		modified bool
		head     *logEntry
		current  *logEntry
		rx       *regexp.Regexp
	}

	logEntry struct {
		line string
		next *logEntry
	}
)

func newLogBuffer(c int, f bool) *logBuffer {
	return &logBuffer{capacity: c, decorate: f, rx: regexp.MustCompile(`\[\w*\:\:\]`)}
}

func (b *logBuffer) clear() {
	b.head, b.current = nil, nil
}

func (b *logBuffer) add(line string) {
	b.modified = true
	if b.decorate {
		line = b.decorateLine(line)
	}
	n := logEntry{line: line}
	if b.head == nil {
		b.head = &n
		b.current = b.head
		return
	}

	if b.full() {
		b.head = b.head.next
	}
	b.current.next = &n
	b.current = &n
}

func (b *logBuffer) full() bool {
	return b.length() == b.capacity
}

func (b *logBuffer) length() int {
	c, count := b.head, 0
	for c != nil {
		c = c.next
		count++
	}
	return count
}

func (*logBuffer) decorateLine(l string) string {
	return "[" + newLogColor + "::]" + l + "[::]"
}

func (b *logBuffer) trimLine(l string) string {
	return b.rx.ReplaceAllString(l, "")
}

func (b *logBuffer) cleanse() {
	if !b.modified {
		return
	}
	c := b.head
	for c != nil {
		c.line = b.trimLine(c.line)
		c = c.next
	}
	b.modified = true
}

func (b *logBuffer) String() string {
	return strings.Join(b.lines(), "\n")
}

func (b *logBuffer) lines() []string {
	out := make([]string, b.length())
	c := b.head
	for i := 0; c != nil; i++ {
		out[i] = c.line
		c = c.next
	}
	return out
}

func (b *logBuffer) dump() {
	c := b.head
	for c != nil {
		fmt.Println(c.line)
		c = c.next
	}
}

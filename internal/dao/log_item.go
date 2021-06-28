package dao

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/derailed/k9s/internal/color"
)

// LogChan represents a channel for logs.
type LogChan chan *LogItem

// LogItem represents a container log line.
type LogItem struct {
	Pod, Container, Timestamp string
	SingleContainer           bool
	Bytes                     []byte
}

// NewLogItem returns a new item.
func NewLogItem(b []byte) *LogItem {
	space := []byte(" ")
	cols := bytes.Split(b[:len(b)-1], space)

	return &LogItem{
		Timestamp: string(cols[0]),
		Bytes:     bytes.Join(cols[1:], space),
	}
}

// NewLogItemFromString returns a new item.
func NewLogItemFromString(s string) *LogItem {
	return &LogItem{
		Bytes:     []byte(s),
		Timestamp: time.Now().String(),
	}
}

// ID returns pod and or container based id.
func (l *LogItem) ID() string {
	if l.Pod != "" {
		return l.Pod
	}
	return l.Container
}

// Clone copies an item.
func (l *LogItem) Clone() *LogItem {
	bytes := make([]byte, len(l.Bytes))
	copy(bytes, l.Bytes)
	return &LogItem{
		Container:       l.Container,
		Pod:             l.Pod,
		Timestamp:       l.Timestamp,
		SingleContainer: l.SingleContainer,
		Bytes:           bytes,
	}
}

// Info returns pod and container information.
func (l *LogItem) Info() string {
	return fmt.Sprintf("%q::%q", l.Pod, l.Container)
}

// IsEmpty checks if the entry is empty.
func (l *LogItem) IsEmpty() bool {
	return len(l.Bytes) == 0
}

var (
	escPattern = regexp.MustCompile(`(\[[a-zA-Z0-9_,;: \-\."#]+\[*)\]`)
	matcher    = []byte("$1[]")
)

// Render returns a log line as string.
func (l *LogItem) Render(paint int, showTime bool) []byte {
	bb := make([]byte, 0, 200)
	if showTime {
		t := l.Timestamp
		for i := len(t); i < 30; i++ {
			t += " "
		}
		bb = append(bb, color.ANSIColorize(t, 106)...)
		bb = append(bb, ' ')
	}

	var hasPod bool
	if l.Pod != "" {
		bb = append(bb, color.ANSIColorize(l.Pod, paint)...)
		hasPod = true
	}
	if !l.SingleContainer && l.Container != "" {
		if hasPod {
			bb = append(bb, ':')
		}
		bb = append(bb, color.ANSIColorize(l.Container, paint)...)
		bb = append(bb, ' ')
	} else if hasPod {
		bb = append(bb, ' ')
	}

	return append(bb, escPattern.ReplaceAll(l.Bytes, matcher)...)
}

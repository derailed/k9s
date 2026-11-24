// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"

	"github.com/derailed/tview"
)

// LogChan represents a channel for logs.
type LogChan chan *LogItem

var ItemEOF = new(LogItem)

// RenderOpts tracks how log lines are rendered.
type RenderOpts struct {
	// ShowTime displays the log line timestamp.
	ShowTime bool

	// PrettyJSON indents log lines holding a JSON payload.
	PrettyJSON bool
}

// LogItem represents a container log line.
type LogItem struct {
	Pod, Container  string
	SingleContainer bool
	Bytes           []byte
	IsError         bool

	// rawJSON holds the raw (unescaped) message when it looks like JSON.
	rawJSON []byte
}

// NewLogItem returns a new item.
func NewLogItem(bb []byte) *LogItem {
	return &LogItem{
		Bytes: bb,
	}
}

// NewRawLogItem returns a new item from a raw log line. The line is escaped
// for display and, when its message looks like JSON, the raw message is kept
// so that it can be pretty printed.
func NewRawLogItem(raw []byte) *LogItem {
	item := NewLogItem(tview.EscapeBytes(raw))
	msg := raw
	if index := bytes.IndexByte(raw, ' '); index > 0 {
		msg = raw[index+1:]
	}
	if maybeJSON(bytes.TrimSpace(msg)) {
		item.rawJSON = msg
	}

	return item
}

// NewLogItemFromString returns a new item.
func NewLogItemFromString(s string) *LogItem {
	return &LogItem{
		Bytes: []byte(s),
	}
}

// ID returns pod and or container based id.
func (l *LogItem) ID() string {
	if l.Pod != "" {
		return l.Pod
	}
	return l.Container
}

// GetTimestamp fetch log lime timestamp
func (l *LogItem) GetTimestamp() string {
	index := bytes.Index(l.Bytes, []byte{' '})
	if index < 0 {
		return ""
	}
	return string(l.Bytes[:index])
}

// Info returns pod and container information.
func (l *LogItem) Info() string {
	return l.Pod + "::" + l.Container
}

// IsEmpty checks if the entry is empty.
func (l *LogItem) IsEmpty() bool {
	return len(l.Bytes) == 0
}

// Size returns the size of the item.
func (l *LogItem) Size() int {
	return 100 + len(l.Bytes) + len(l.Pod) + len(l.Container)
}

// Render returns a log line as string.
func (l *LogItem) Render(paint string, opts RenderOpts, bb *bytes.Buffer) {
	index := bytes.Index(l.Bytes, []byte{' '})
	if opts.ShowTime && index > 0 {
		bb.WriteString("[gray::b]")
		bb.Write(l.Bytes[:index])
		bb.WriteString(" ")
		if l := 30 - len(l.Bytes[:index]); l > 0 {
			bb.Write(bytes.Repeat([]byte{' '}, l))
		}
		bb.WriteString("[-::-]")
	}

	if l.Pod != "" {
		bb.WriteString("[" + paint + "::]" + l.Pod)
	}

	if !l.SingleContainer && l.Container != "" {
		if l.Pod != "" {
			bb.WriteString(" ")
		}
		bb.WriteString("[" + paint + "::b]" + l.Container + "[-::-] ")
	} else if l.Pod != "" {
		bb.WriteString("[-::] ")
	}

	if opts.PrettyJSON && l.rawJSON != nil {
		if pj, ok := prettyJSON(l.rawJSON); ok {
			bb.Write(tview.EscapeBytes(pj))
			return
		}
	}

	if index > 0 {
		bb.Write(l.Bytes[index+1:])
	} else {
		bb.Write(l.Bytes)
	}
}

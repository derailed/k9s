package dao

import (
	"bytes"
	"regexp"
	"strings"
)

// LogChan represents a channel for logs.
type LogChan chan *LogItem

var ItemEOF = new(LogItem)

// LogItem represents a container log line.
type LogItem struct {
	Pod, Container  string
	SingleContainer bool
	Bytes           []byte
	IsError         bool
}

// NewLogItem returns a new item.
func NewLogItem(bb []byte) *LogItem {
	return &LogItem{
		Bytes: bb,
	}
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

// TODO: make constant
func ansiRegex() *regexp.Regexp {
	// referenced
	// - https://github.com/chalk/ansi-regex/blob/main/index.js
	// - https://github.com/acarl005/stripansi/blob/master/stripansi.go
	reg := []string{
		"[\u001B\u009B][[\\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]+)*|[a-zA-Z\\d]+(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\u0007)",
		"(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-nq-uy=><~]))",
	}
	pattern := strings.Join(reg[:], "|")
	return regexp.MustCompile(pattern)
}

func stripAnsi(b []byte) []byte {
	return []byte(ansiRegex().ReplaceAllString(string(b), ""))
}

// Render returns a log line as string.
func (l *LogItem) Render(paint string, showTime bool, bb *bytes.Buffer) {
	index := bytes.Index(l.Bytes, []byte{' '})
	if showTime && index > 0 {
		bb.WriteString("[gray::b]")
		bb.Write(l.Bytes[:index])
		bb.WriteString(" ")
		for i := len(l.Bytes[:index]); i < 30; i++ {
			bb.WriteByte(' ')
		}
		bb.WriteString("[-::]")
	}

	if l.Pod != "" {
		bb.WriteString("[" + paint + "::]" + l.Pod)
	}

	if !l.SingleContainer && l.Container != "" {
		if len(l.Pod) > 0 {
			bb.WriteString(" ")
		}
		bb.WriteString("[" + paint + "::b]" + l.Container + "[-::-] ")
	} else if len(l.Pod) > 0 {
		bb.WriteString("[-::] ")
	}

	// TODO: make configurable k9s.logger.stripAnsi
	shouldStripAnsi := true

	if index > 0 {
		if shouldStripAnsi {
			bb.Write(stripAnsi(l.Bytes[index+1:]))
		} else {
			bb.Write(l.Bytes[index+1:])
		}
	} else {
		if shouldStripAnsi {
			bb.Write(stripAnsi(l.Bytes))
		} else {
			bb.Write(l.Bytes)
		}
	}
}

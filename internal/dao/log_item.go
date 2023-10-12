// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"os"
	"time"

	"github.com/rs/zerolog/log"
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

// Render returns a log line as string.
func (l *LogItem) Render(paint string, showTime bool, bb *bytes.Buffer) {
	index := bytes.Index(l.Bytes, []byte{' '})

	if showTime && index > 0 {
		bb.WriteString("[gray::b]")
		date := l.Bytes[:index]

		if os.Getenv("TZ") != "" {
			t, err := time.Parse("2006-01-02T15:04:05.000000000Z", string(date))

			if err != nil {
				log.Error().Err(err).Msg("Invalid date log format")
				return
			}

			loc, err := time.LoadLocation(os.Getenv("TZ"))

			if err != nil {
				log.Error().Err(err).Msg("Invalid timezone")
				return
			}

			dateTimezoned := t.In(loc).Format("2006-01-02T15:04:05.000000000Z")

			bb.Write([]byte(dateTimezoned))
		} else {
			bb.Write(date)
		}

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
		if len(l.Pod) > 0 {
			bb.WriteString(" ")
		}
		bb.WriteString("[" + paint + "::b]" + l.Container + "[-::-] ")
	} else if len(l.Pod) > 0 {
		bb.WriteString("[-::] ")
	}

	if index > 0 {
		bb.Write(l.Bytes[index+1:])
	} else {
		bb.Write(l.Bytes)
	}
}

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

var locationFromTZ *time.Location = nil

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

// RenderTimestampWithTimezone transform the date to the desired timezone and write the date in bb.
func (l *LogItem) RenderTimestampWithTimezone(date []byte, bb *bytes.Buffer) {
	t, err := time.Parse("2006-01-02T15:04:05.000000000Z", string(date))

	if err != nil {
		log.Error().Err(err).Msg("Invalid date log format")
		return
	}

	if locationFromTZ == nil {
		locationFromTZ, err = time.LoadLocation(os.Getenv("TZ"))

		if err != nil {
			log.Error().Err(err).Msg("Invalid timezone")
			return
		}
	}

	bb.Write([]byte(t.In(locationFromTZ).Format("2006-01-02T15:04:05.000000000Z")))
}

// RenderTimestamp write the date in bb.
func (l *LogItem) RenderTimestamp(date []byte, index int, bb *bytes.Buffer) {
	bb.WriteString("[gray::b]")

	if os.Getenv("TZ") != "" {
		l.RenderTimestampWithTimezone(date, bb)
	} else {
		bb.Write(date)
	}

	bb.WriteString(" ")
	for i := len(l.Bytes[:index]); i < 30; i++ {
		bb.WriteByte(' ')
	}
	bb.WriteString("[-::-]")
}

// Render returns a log line as string.
func (l *LogItem) Render(paint string, showTime bool, bb *bytes.Buffer) {
	index := bytes.Index(l.Bytes, []byte{' '})

	if showTime && index > 0 {
		date := l.Bytes[:index]
		l.RenderTimestamp(date, index, bb)
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

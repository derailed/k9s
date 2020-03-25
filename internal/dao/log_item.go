package dao

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

// LogChan represents a channel for logs.
type LogChan chan *LogItem

// LogItem represents a container log line.
type LogItem struct {
	Pod, Container, Timestamp string
	Bytes                     []byte
}

// NewLogItem returns a new item.
func NewLogItem(b []byte) *LogItem {
	space := []byte(" ")
	var l LogItem

	cols := bytes.Split(b[:len(b)-1], space)
	l.Timestamp = string(cols[0])
	l.Bytes = bytes.Join(cols[1:], space)

	return &l
}

// NewLogItemFromString returns a new item.
func NewLogItemFromString(s string) *LogItem {
	l := LogItem{Bytes: []byte(s)}
	l.Timestamp = time.Now().String()

	return &l
}

// ID returns pod and or container based id.
func (l *LogItem) ID() string {
	if l.Pod != "" {
		return l.Pod
	}
	return l.Container
}

// Info returns pod and container information.
func (l *LogItem) Info() string {
	return fmt.Sprintf("%q::%q", l.Pod, l.Container)
}

// IsEmpty checks if the entry is empty.
func (l *LogItem) IsEmpty() bool {
	return len(l.Bytes) == 0
}

const colorFmt = "\033[38;5;%dm%s\033[0m"

// colorize me
func colorize(s string, c int) string {
	return fmt.Sprintf(colorFmt, c, s)
}

// Render returns a log line as string.
func (l *LogItem) Render(c int, showTime bool) []byte {
	bb := make([]byte, 0, 30+len(l.Bytes)+len(l.Info()))
	if showTime {
		bb = append(bb, colorize(fmt.Sprintf("%-30s ", l.Timestamp), 106)...)
	}

	if l.Pod != "" {
		bb = append(bb, []byte(colorize(l.Pod, c))...)
		bb = append(bb, ':')
	}
	if l.Container != "" {
		bb = append(bb, []byte(colorize(l.Container, c))...)
		bb = append(bb, ' ')
	}
	bb = append(bb, []byte(tview.Escape(string(l.Bytes)))...)

	return bb
}

func colorFor(n string) int {
	var sum int
	for _, r := range n {
		sum += int(r)
	}

	c := sum % 256
	if c == 0 {
		c = 207 + rand.Intn(10)
	}
	return c
}

// ----------------------------------------------------------------------------

// LogItems represents a collection of log items.
type LogItems []*LogItem

// Lines returns a collection of log lines.
func (l LogItems) Lines() []string {
	ll := make([]string, len(l))
	for i, item := range l {
		ll[i] = string(item.Render(0, false))
	}

	return ll
}

// Render returns logs as a collection of strings.
func (l LogItems) Render(showTime bool, ll [][]byte) {
	colors := map[string]int{}
	for i, item := range l {
		info := item.ID()
		c, ok := colors[item.ID()]
		if !ok {
			c = colorFor(info)
			colors[info] = c
		}
		ll[i] = item.Render(c, showTime)
	}
}

// DumpDebug for debuging
func (l LogItems) DumpDebug(m string) {
	fmt.Println(m + strings.Repeat("-", 50))
	for i, line := range l {
		fmt.Println(i, string(line.Bytes))
	}
}

// Filter filters out log items based on given filter.
func (l LogItems) Filter(q string) ([]int, error) {
	if q == "" {
		return nil, nil
	}
	if IsFuzzySelector(q) {
		return l.fuzzyFilter(strings.TrimSpace(q[2:])), nil
	}
	indexes, err := l.filterLogs(q)
	if err != nil {
		log.Error().Err(err).Msgf("Logs filter failed")
		return nil, err
	}
	return indexes, nil
}

var fuzzyRx = regexp.MustCompile(`\A\-f`)

func (l LogItems) fuzzyFilter(q string) []int {
	q = strings.TrimSpace(q)
	matches := make([]int, 0, len(l))
	mm := fuzzy.Find(q, l.Lines())
	for _, m := range mm {
		matches = append(matches, m.Index)
	}

	return matches
}

func (l LogItems) filterLogs(q string) ([]int, error) {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil, err
	}
	matches := make([]int, 0, len(l))
	for i, line := range l.Lines() {
		if rx.MatchString(line) {
			matches = append(matches, i)
		}
	}

	return matches, nil
}

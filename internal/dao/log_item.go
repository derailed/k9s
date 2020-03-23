package dao

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

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

// IsEmpty checks if the entry is empty.
func (l *LogItem) IsEmpty() bool {
	return len(l.Bytes) == 0
}

// Render returns a log line as string.
func (l *LogItem) Render(showTime bool) []byte {
	bb := make([]byte, 0, 100+len(l.Bytes))
	if showTime {
		bb = append(bb, fmt.Sprintf("%-30s ", l.Timestamp)...)
	}

	if l.Pod != "" {
		bb = append(bb, l.Pod...)
		bb = append(bb, ':')
		bb = append(bb, l.Container...)
		bb = append(bb, ' ')
	} else if l.Container != "" {
		bb = append(bb, l.Container...)
		bb = append(bb, ' ')
	}
	bb = append(bb, l.Bytes...)

	return bb
}

// ----------------------------------------------------------------------------

// LogItems represents a collection of log items.
type LogItems []*LogItem

// Lines returns a collection of log lines.
func (l LogItems) Lines() []string {
	ll := make([]string, len(l))
	for i, item := range l {
		ll[i] = string(item.Render(false))
	}

	return ll
}

// Render returns logs as a collection of strings.
func (l LogItems) Render(showTime bool, ll [][]byte) {
	for i, item := range l {
		ll[i] = item.Render(showTime)
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

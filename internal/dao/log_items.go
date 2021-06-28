package dao

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/sahilm/fuzzy"
)

var colorPalette = []tcell.Color{
	tcell.ColorTeal,
	tcell.ColorGreen,
	tcell.ColorPurple,
	tcell.ColorLime,
	tcell.ColorBlue,
	tcell.ColorYellow,
	tcell.ColorFuchsia,
	tcell.ColorAqua,
}

// LogItems represents a collection of log items.
type LogItems struct {
	items  []*LogItem
	colors map[string]tcell.Color
	mx     sync.RWMutex
}

// NewLogItems returns a new instance.
func NewLogItems() *LogItems {
	return &LogItems{
		colors: make(map[string]tcell.Color),
	}
}

// Len returns the items length.
func (l *LogItems) Items() []*LogItem {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.items
}

// Len returns the items length.
func (l *LogItems) Len() int {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return len(l.items)
}

// Clear removes all items.
func (l *LogItems) Clear() {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = nil
	for k := range l.colors {
		delete(l.colors, k)
	}
}

// Shift scrolls the lines by one.
func (l *LogItems) Shift(i *LogItem) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = append(l.items[1:], i)
}

// Subset return a subset of logitems.
func (l *LogItems) Subset(index int) *LogItems {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return &LogItems{
		items:  l.items[index:],
		colors: l.colors,
	}
}

// Merge merges two logitems list.
func (l *LogItems) Merge(n *LogItems) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = append(l.items, n.items...)
	for k, v := range n.colors {
		l.colors[k] = v
	}
}

// Add augments the items.
func (l *LogItems) Add(ii ...*LogItem) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = append(l.items, ii...)
}

// Lines returns a collection of log lines.
func (l *LogItems) Lines(showTime bool) [][]byte {
	l.mx.Lock()
	defer l.mx.Unlock()

	ll := make([][]byte, len(l.items))
	for i, item := range l.items {
		color := l.colors[item.ID()]
		ll[i] = item.Render(int(color-tcell.ColorValid), showTime)
	}

	return ll
}

// StrLines returns a collection of log lines.
func (l *LogItems) StrLines(showTime bool) []string {
	l.mx.Lock()
	defer l.mx.Unlock()

	ll := make([]string, len(l.items))
	for i, item := range l.items {
		ll[i] = string(item.Render(0, showTime))
	}

	return ll
}

// Render returns logs as a collection of strings.
func (l *LogItems) Render(showTime bool, ll [][]byte) {
	index := len(l.colors)
	for i, item := range l.items {
		id := item.ID()
		color, ok := l.colors[id]
		if !ok {
			if index >= len(colorPalette) {
				index = 0
			}
			color = colorPalette[index]
			l.colors[id] = color
			index++
		}
		ll[i] = item.Render(int(color-tcell.ColorValid), showTime)
	}
}

// DumpDebug for debugging
func (l *LogItems) DumpDebug(m string) {
	fmt.Println(m + strings.Repeat("-", 50))
	for i, line := range l.items {
		fmt.Println(i, string(line.Bytes))
	}
}

// Filter filters out log items based on given filter.
func (l *LogItems) Filter(q string, showTime bool) ([]int, [][]int, error) {
	if q == "" {
		return nil, nil, nil
	}
	if IsFuzzySelector(q) {
		mm, ii := l.fuzzyFilter(strings.TrimSpace(q[2:]), showTime)
		return mm, ii, nil
	}
	matches, indices, err := l.filterLogs(q, showTime)
	if err != nil {
		return nil, nil, err
	}

	return matches, indices, nil
}

func (l *LogItems) fuzzyFilter(q string, showTime bool) ([]int, [][]int) {
	q = strings.TrimSpace(q)
	matches, indices := make([]int, 0, len(l.items)), make([][]int, 0, 10)
	mm := fuzzy.Find(q, l.StrLines(showTime))
	for _, m := range mm {
		matches = append(matches, m.Index)
		indices = append(indices, m.MatchedIndexes)
	}

	return matches, indices
}

func (l *LogItems) filterLogs(q string, showTime bool) ([]int, [][]int, error) {
	var invert bool
	if IsInverseSelector(q) {
		invert = true
		q = q[1:]
	}
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil, nil, err
	}
	matches, indices := make([]int, 0, len(l.items)), make([][]int, 0, 10)
	for i, line := range l.Lines(showTime) {
		locs := rx.FindIndex(line)
		if locs != nil && invert {
			continue
		}
		if locs == nil && !invert {
			continue
		}
		matches = append(matches, i)
		ii := make([]int, 0, 10)
		for i := 0; i < len(locs); i += 2 {
			for j := locs[i]; j < locs[i+1]; j++ {
				ii = append(ii, j)
			}
		}
		indices = append(indices, ii)
	}

	return matches, indices, nil
}

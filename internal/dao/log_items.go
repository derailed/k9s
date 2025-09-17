// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal"
	"github.com/sahilm/fuzzy"
)

type podColors map[string]string

var podPalette = []string{
	"teal",
	"green",
	"purple",
	"lime",
	"blue",
	"yellow",
	"fushia",
	"aqua",
}

// LogItems represents a collection of log items.
type LogItems struct {
	items     []*LogItem
	podColors podColors
	mx        sync.RWMutex
}

// NewLogItems returns a new instance.
func NewLogItems() *LogItems {
	return &LogItems{
		podColors: make(map[string]string),
	}
}

// Items returns the log items.
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

	l.items = l.items[:0]
	for k := range l.podColors {
		delete(l.podColors, k)
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
		items:     l.items[index:],
		podColors: l.podColors,
	}
}

// Merge merges two logitems list.
func (l *LogItems) Merge(n *LogItems) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = append(l.items, n.items...)
	for k, v := range n.podColors {
		l.podColors[k] = v
	}
}

// Add augments the items.
func (l *LogItems) Add(ii ...*LogItem) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.items = append(l.items, ii...)
}

func (l *LogItems) podColorFor(id string) string {
	color, ok := l.podColors[id]
	if ok {
		return color
	}
	var idx int
	for i, r := range id {
		idx += i * int(r)
	}
	l.podColors[id] = podPalette[idx%len(podPalette)]

	return l.podColors[id]
}

// Lines returns a collection of log lines.
func (l *LogItems) Lines(index int, showTime bool, ll [][]byte) {
	l.mx.Lock()
	defer l.mx.Unlock()

	for i, item := range l.items[index:] {
		bb := bytes.NewBuffer(make([]byte, 0, item.Size()))
		item.Render(l.podColorFor(item.ID()), showTime, bb)
		ll[i] = bb.Bytes()
	}
}

// StrLines returns a collection of log lines.
func (l *LogItems) StrLines(index int, showTime bool) []string {
	l.mx.Lock()
	defer l.mx.Unlock()

	ll := make([]string, len(l.items[index:]))
	for i, item := range l.items[index:] {
		bb := bytes.NewBuffer(make([]byte, 0, item.Size()))
		item.Render(l.podColorFor(item.ID()), showTime, bb)
		ll[i] = bb.String()
	}

	return ll
}

// Render returns logs as a collection of strings.
func (l *LogItems) Render(index int, showTime bool, ll [][]byte) {
	for i, item := range l.items[index:] {
		bb := bytes.NewBuffer(make([]byte, 0, item.Size()))
		item.Render(l.podColorFor(item.ID()), showTime, bb)
		ll[i] = bb.Bytes()
	}
}

// DumpDebug for debugging.
func (l *LogItems) DumpDebug(m string) {
	fmt.Println(m + strings.Repeat("-", 50))
	for i, line := range l.items {
		fmt.Println(i, string(line.Bytes))
	}
}

// Filter filters out log items based on given filter.
func (l *LogItems) Filter(index int, q string, showTime bool) (matches []int, indices [][]int, err error) {
	if q == "" {
		return
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		matches, indices = l.fuzzyFilter(index, f, showTime)
		return
	}
	matches, indices, err = l.filterLogs(index, q, showTime)
	if err != nil {
		return
	}

	return matches, indices, nil
}

func (l *LogItems) fuzzyFilter(index int, q string, showTime bool) (matches []int, indices [][]int) {
	q = strings.TrimSpace(q)
	matches, indices = make([]int, 0, len(l.items)), make([][]int, 0, len(l.items))
	mm := fuzzy.Find(q, l.StrLines(index, showTime))
	for _, m := range mm {
		matches = append(matches, m.Index)
		indices = append(indices, m.MatchedIndexes)
	}

	return matches, indices
}

func (l *LogItems) filterLogs(index int, q string, showTime bool) (matches []int, indices [][]int, err error) {
	var invert bool
	if internal.IsInverseSelector(q) {
		invert = true
		q = q[1:]
	}
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil, nil, err
	}
	matches, indices = make([]int, 0, len(l.items)), make([][]int, 0, len(l.items))
	ll := make([][]byte, len(l.items[index:]))
	l.Lines(index, showTime, ll)
	for i, line := range ll {
		locs := rx.FindAllIndex(line, -1)
		if locs != nil && invert {
			continue
		}
		if locs == nil && !invert {
			continue
		}
		matches = append(matches, i)
		ii := make([]int, 0, 10)
		for _, loc := range locs {
			for j := loc[0]; j < loc[1]; j++ {
				ii = append(ii, j)
			}
		}
		indices = append(indices, ii)
	}

	return matches, indices, nil
}

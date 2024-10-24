// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"fmt"
	"image"
	"math"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type block struct {
	full    int
	partial rune
}

type blocks struct {
	s1, s2 block
}

// Metric tracks two series.
type Metric struct {
	S1, S2 int64
}

// MaxDigits returns the max series number of digits.
func (m Metric) MaxDigits() int {
	s := fmt.Sprintf("%d", m.Max())

	return len(s)
}

// Max returns the max of the series.
func (m Metric) Max() int64 {
	return int64(math.Max(float64(m.S1), float64(m.S2)))
}

// Sum returns the sum of series.
func (m Metric) Sum() int64 {
	return m.S1 + m.S2
}

// SparkLine represents a sparkline component.
type SparkLine struct {
	*Component

	data        []Metric
	multiSeries bool
}

// NewSparkLine returns a new graph.
func NewSparkLine(id string) *SparkLine {
	return &SparkLine{
		Component:   NewComponent(id),
		multiSeries: true,
	}
}

// SetMultiSeries indicates if multi series are in effect or not.
func (s *SparkLine) SetMultiSeries(b bool) {
	s.multiSeries = b
}

// Add adds a metric.
func (s *SparkLine) Add(m Metric) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.data = append(s.data, m)
}

// Draw draws the graph.
func (s *SparkLine) Draw(screen tcell.Screen) {
	s.Component.Draw(screen)

	s.mx.RLock()
	defer s.mx.RUnlock()

	if len(s.data) == 0 {
		return
	}

	pad := 0
	if s.legend != "" {
		pad++
	}

	rect := s.asRect()
	s.cutSet(rect.Dx())
	max := s.computeMax()

	cX, idx := rect.Min.X+1, 0
	if len(s.data)*2 < rect.Dx() {
		cX = rect.Max.X - len(s.data)*2
	} else {
		idx = len(s.data) - rect.Dx()/2
	}

	scale := float64(len(sparks)*(rect.Dy()-pad)) / float64(max)
	c1, c2 := s.colorForSeries()
	for _, d := range s.data[idx:] {
		b := toBlocks(d, scale)
		cY := rect.Max.Y - pad
		s.drawBlock(rect, screen, cX, cY, b.s1, c1)
		cX++
		s.drawBlock(rect, screen, cX, cY, b.s2, c2)
		cX++
	}

	if rect.Dx() > 0 && rect.Dy() > 0 && s.legend != "" {
		legend := s.legend
		if s.HasFocus() {
			legend = fmt.Sprintf("[%s:%s:]", s.focusFgColor, s.focusBgColor) + s.legend + "[::]"
		}
		tview.Print(screen, legend, rect.Min.X, rect.Max.Y, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (s *SparkLine) drawBlock(r image.Rectangle, screen tcell.Screen, x, y int, b block, c tcell.Color) {
	style := tcell.StyleDefault.Foreground(c).Background(s.bgColor)

	zeroY := r.Max.Y - r.Dy()
	for i := 0; i < b.full; i++ {
		screen.SetContent(x, y, sparks[len(sparks)-1], nil, style)
		y--
		if y <= zeroY {
			break
		}
	}
	if b.partial != 0 {
		screen.SetContent(x, y, b.partial, nil, style)
	}
}

func (s *SparkLine) cutSet(width int) {
	if width <= 0 || len(s.data) == 0 {
		return
	}

	if len(s.data) >= width*2 {
		s.data = s.data[len(s.data)-width:]
	}
}

func (s *SparkLine) computeMax() int64 {
	var max int64
	for _, d := range s.data {
		m := d.Max()
		if max < m {
			max = m
		}
	}

	return max
}

func toBlocks(m Metric, scale float64) blocks {
	if m.Sum() <= 0 {
		return blocks{}
	}
	return blocks{s1: makeBlocks(m.S1, scale), s2: makeBlocks(m.S2, scale)}
}

func makeBlocks(v int64, scale float64) block {
	scaled := int(math.Round(float64(v) * scale))
	p, b := scaled%len(sparks), block{full: scaled / len(sparks)}
	if b.full == 0 && v > 0 && p == 0 {
		p = 4
	}
	if v > 0 && p >= 0 && p < len(sparks) {
		b.partial = sparks[p]
	}

	return b
}

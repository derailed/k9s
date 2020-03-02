package tchart

import (
	"fmt"
	"image"
	"math"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type block struct {
	full    int
	partial rune
}

type blocks struct {
	oks, errs block
}

// Metric tracks a good and error rates.
type Metric struct {
	OK, Fault int
}

// MaxDigits returns the max series number of digits.
func (m Metric) MaxDigits() int {

	s := fmt.Sprintf("%d", m.Max())
	return len(s)
}

// Max returns the max of the series.
func (m Metric) Max() int {
	return int(math.Max(float64(m.OK), float64(m.Fault)))
}

// Sum returns the sum of the metrics.
func (m Metric) Sum() int {
	return m.OK + m.Fault
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

	pad := 1
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

	factor := 2
	if !s.multiSeries {
		factor = 1
	}
	scale := float64(len(sparks)*(rect.Dy()-pad)/factor) / float64(max)
	c1, c2 := s.colorForSeries()
	for _, d := range s.data[idx:] {
		b := toBlocks(d, scale)
		cY := rect.Max.Y - pad
		cY = s.drawBlock(rect, screen, cX, cY, b.oks, c1)
		_ = s.drawBlock(rect, screen, cX, cY, b.errs, c2)
		cX += 2
	}

	if rect.Dx() > 0 && rect.Dy() > 0 && s.legend != "" {
		legend := s.legend
		if s.HasFocus() {
			legend = "[:aqua:]" + s.legend + "[::]"
		}
		tview.Print(screen, legend, rect.Min.X, rect.Max.Y, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (s *SparkLine) drawBlock(r image.Rectangle, screen tcell.Screen, x, y int, b block, c tcell.Color) int {
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
		if b.full == 0 {
			y--
		}
	}

	return y
}

func (s *SparkLine) cutSet(width int) {
	if width <= 0 || len(s.data) == 0 {
		return
	}

	if len(s.data) >= width*2 {
		s.data = s.data[len(s.data)-width:]
	}
}

func (s *SparkLine) computeMax() int {
	var max int
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
	return blocks{oks: makeBlocks(m.OK, scale), errs: makeBlocks(m.Fault, scale)}
}

func makeBlocks(v int, scale float64) block {
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

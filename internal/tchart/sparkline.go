package tchart

import (
	"fmt"
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

// Max returns the max of the metric.
func (m Metric) MaxDigits() int {
	max := int(math.Max(float64(m.OK), float64(m.Fault)))
	s := fmt.Sprintf("%d", max)
	return len(s)
}

// Sum returns the sum of the metrics.
func (m Metric) Sum() int {
	return m.OK + m.Fault
}

// Sparkline represents a sparkline component.
type SparkLine struct {
	*Component

	data      []Metric
	lastWidth int
}

// NewSparkLine returns a new graph.
func NewSparkLine(id string) *SparkLine {
	return &SparkLine{
		Component: NewComponent(id),
	}
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
	s.lastWidth = rect.Dx()
	s.cutSet(rect.Dx())
	max := s.computeMax()

	cX := rect.Min.X + 1
	if len(s.data) < rect.Dx() {
		cX = rect.Max.X - len(s.data)
	}

	scale := float64(len(sparks)) * float64((rect.Dy() - pad)) / float64(max)

	c1, c2 := s.colorForSeries()
	for _, d := range s.data {
		b := toBlocks(d, scale)
		cY := rect.Max.Y - pad
		cY = s.drawBlock(screen, cX, cY, b.oks, c1)
		s.drawBlock(screen, cX, cY, b.errs, c2)
		cX++
	}

	if rect.Dx() > 0 && rect.Dy() > 0 && s.legend != "" {
		legend := s.legend
		if s.HasFocus() {
			legend = "[:aqua:]" + s.legend + "[::]"
		}
		tview.Print(screen, legend, rect.Min.X, rect.Max.Y-1, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (s *SparkLine) drawBlock(screen tcell.Screen, x, y int, b block, c tcell.Color) int {
	style := tcell.StyleDefault.Foreground(c).Background(s.bgColor)

	for i := 0; i < b.full; i++ {
		screen.SetContent(x, y, sparks[len(sparks)-1], nil, style)
		y--
	}
	if b.partial != 0 {
		screen.SetContent(x, y, b.partial, nil, style)
	}

	return y
}

func (s *SparkLine) cutSet(w int) {
	if w <= 0 || len(s.data) == 0 {
		return
	}

	if w < len(s.data) {
		s.data = s.data[len(s.data)-w:]
	}
}

func (s *SparkLine) computeMax() int {
	var max int
	for _, d := range s.data {
		sum := d.Sum()
		if sum > max {
			max = sum
		}
	}

	return max
}

func toBlocks(value Metric, scale float64) blocks {
	if value.Sum() <= 0 {
		return blocks{}
	}

	oks := int(math.Floor(float64(value.OK) * scale))
	part, okB := oks%len(sparks), block{full: oks / len(sparks)}
	if part > 0 {
		okB.partial = sparks[part-1]
	}

	errs := int(math.Round(float64(value.Fault) * scale))
	part, errB := errs%len(sparks), block{full: errs / len(sparks)}
	if part > 0 {
		errB.partial = sparks[part-1]
	}

	return blocks{oks: okB, errs: errB}
}

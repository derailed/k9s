package tchart

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

const axisColor = "#ff0066"

type block struct {
	full    int
	partial rune
}

// SparkLine represents a sparkline component.
type SparkLine struct {
	*Component

	series     MetricSeries
	max        float64
	unit       string
	colorIndex int
}

// NewSparkLine returns a new graph.
func NewSparkLine(id, unit string) *SparkLine {
	return &SparkLine{
		Component: NewComponent(id),
		series:    make(MetricSeries),
		unit:      unit,
	}
}

// GetSeriesColorNames returns series colors by name.
func (s *SparkLine) GetSeriesColorNames() []string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	nn := make([]string, 0, len(s.seriesColors))
	for _, color := range s.seriesColors {
		for name, co := range tcell.ColorNames {
			if co == color {
				nn = append(nn, name)
			}
		}
	}
	if len(nn) < 3 {
		nn = append(nn, "green", "orange", "orangered")
	}

	return nn
}

func (s *SparkLine) SetColorIndex(n int) {
	s.colorIndex = n
}

func (s *SparkLine) SetMax(m float64) {
	if m > s.max {
		s.max = m
	}
}

func (s *SparkLine) GetMax() float64 {
	return s.max
}

func (*SparkLine) Add(int, int) {}

// Add adds a metric.
func (s *SparkLine) AddMetric(t time.Time, f float64) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.series.Add(t, f)
}

func (s *SparkLine) printYAxis(screen tcell.Screen, rect image.Rectangle) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor(axisColor)).Background(s.bgColor)
	for y := range rect.Dy() - 3 {
		screen.SetContent(rect.Min.X, rect.Min.Y+y, tview.BoxDrawingsLightVertical, nil, style)
	}
	screen.SetContent(rect.Min.X, rect.Min.Y+rect.Dy()-3, tview.BoxDrawingsLightUpAndRight, nil, style)
}

func (s *SparkLine) printXAxis(screen tcell.Screen, rect image.Rectangle) time.Time {
	dx, t := rect.Dx()-1, time.Now()
	vals := make([]string, 0, dx)
	for i := dx; i > 0; i -= 10 {
		label := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
		vals = append(vals, label)
		t = t.Add(-(10 * time.Minute))
	}

	y := rect.Max.Y - 2
	for _, v := range vals {
		if dx <= 2 {
			break
		}
		tview.Print(screen, v, rect.Min.X+dx-5, y, 6, tview.AlignCenter, tcell.ColorOrange)
		dx -= 10
	}
	style := tcell.StyleDefault.Foreground(tcell.GetColor(axisColor)).Background(s.bgColor)
	for x := 1; x < rect.Dx()-1; x++ {
		screen.SetContent(rect.Min.X+x, rect.Max.Y-3, tview.BoxDrawingsLightHorizontal, nil, style)
	}

	return t
}

// Draw draws the graph.
func (s *SparkLine) Draw(screen tcell.Screen) {
	s.Component.Draw(screen)

	s.mx.RLock()
	defer s.mx.RUnlock()

	rect := s.asRect()
	s.printXAxis(screen, rect)

	padX := 1
	s.cutSet(rect.Dx() - padX)
	var cX int
	if len(s.series) < rect.Dx() {
		cX = rect.Max.X - len(s.series) - 1
	} else {
		cX = rect.Min.X + padX
	}

	pad := 2
	if s.legend != "" {
		pad++
	}
	scale := float64(len(sparks)*(rect.Dy()-pad)) / float64(s.max)
	colors := s.colorForSeries()
	cY := rect.Max.Y - pad - 1
	for _, t := range s.series.Keys() {
		b := s.makeBlock(s.series[t], scale)
		s.drawBlock(rect, screen, cX, cY, b, colors[s.colorIndex%len(colors)])
		cX++
	}

	s.printYAxis(screen, rect)

	if rect.Dx() > 0 && rect.Dy() > 0 && s.legend != "" {
		legend := s.legend
		if s.HasFocus() {
			legend = fmt.Sprintf("[%s:%s:]", s.focusFgColor, s.focusBgColor) + s.legend + "[::]"
		}
		tview.Print(screen, legend, rect.Min.X, rect.Max.Y-1, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (s *SparkLine) drawBlock(r image.Rectangle, screen tcell.Screen, x, y int, b block, c tcell.Color) {
	style := tcell.StyleDefault.Foreground(c).Background(s.bgColor)

	zeroY, full := r.Min.Y, sparks[len(sparks)-1]
	for range b.full {
		screen.SetContent(x, y, full, nil, style)
		y--
		if y < zeroY {
			break
		}
	}
	if b.partial != 0 {
		screen.SetContent(x, y, b.partial, nil, style)
	}
}

func (s *SparkLine) cutSet(width int) {
	if width <= 0 || s.series.Empty() {
		return
	}
	if len(s.series) > width {
		s.series.Truncate(width)
	}
}

func (*SparkLine) makeBlock(v, scale float64) block {
	sc := (v * scale)
	scaled := math.Round(sc)
	p, b := int(scaled)%len(sparks), block{full: int(scaled / float64(len(sparks)))}
	if v < 0 {
		return b
	}
	if p > 0 && p < len(sparks) {
		b.partial = sparks[p]
	}

	return b
}

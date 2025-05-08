package tchart

import (
	"fmt"
	"image"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	// DeltaSame represents no difference.
	DeltaSame delta = iota

	// DeltaMore represents a higher value.
	DeltaMore

	// DeltaLess represents a lower value.
	DeltaLess
)

type State struct {
	OK, Fault int
}

type delta int

// Gauge represents a gauge component.
type Gauge struct {
	*Component

	state               State
	resolution          int
	deltaOK, deltaFault delta
}

// NewGauge returns a new gauge.
func NewGauge(id string) *Gauge {
	return &Gauge{
		Component: NewComponent(id),
	}
}

// SetResolution overrides the default number of digits to display.
func (g *Gauge) SetResolution(n int) {
	g.resolution = n
}

// IsDial returns true if chart is a dial
func (*Gauge) IsDial() bool {
	return true
}

func (*Gauge) SetColorIndex(int) {}
func (*Gauge) SetMax(float64)    {}
func (*Gauge) GetMax() float64   { return 0 }

// Add adds a metric.
func (*Gauge) AddMetric(time.Time, float64) {}

// Add adds a new metric.
func (g *Gauge) Add(ok, fault int) {
	g.mx.Lock()
	defer g.mx.Unlock()

	g.deltaOK, g.deltaFault = computeDelta(g.state.OK, ok), computeDelta(g.state.Fault, fault)
	g.state = State{OK: ok, Fault: fault}
}

type number struct {
	ok    bool
	val   int
	str   string
	delta delta
}

// Draw draws the primitive.
func (g *Gauge) Draw(sc tcell.Screen) {
	g.Component.Draw(sc)

	g.mx.RLock()
	defer g.mx.RUnlock()

	rect := g.asRect()
	mid := image.Point{X: rect.Min.X + rect.Dx()/2, Y: rect.Min.Y + rect.Dy()/2 - 1}
	var (
		fmat = "%d"
	)
	d1, d2 := fmt.Sprintf(fmat, g.state.OK), fmt.Sprintf(fmat, g.state.Fault)

	style := tcell.StyleDefault.Background(g.bgColor)

	total := len(d1)*3 + len(d2)*3 + 1
	colors := g.colorForSeries()
	o := image.Point{X: mid.X, Y: mid.Y - 1}
	o.X -= total / 2
	g.drawNum(sc, o, number{ok: true, val: g.state.OK, delta: g.deltaOK, str: d1}, style.Foreground(colors[0]).Dim(false))

	o.X, o.Y = o.X+len(d1)*3, mid.Y
	sc.SetContent(o.X, o.Y, '⠔', nil, style)

	o.X, o.Y = o.X+1, mid.Y-1
	g.drawNum(sc, o, number{ok: false, val: g.state.Fault, delta: g.deltaFault, str: d2}, style.Foreground(colors[1]).Dim(false))

	if rect.Dx() > 0 && rect.Dy() > 0 && g.legend != "" {
		legend := g.legend
		if g.HasFocus() {
			legend = fmt.Sprintf("[%s:%s:]", g.focusFgColor, g.focusBgColor) + g.legend + "[::]"
		}
		tview.Print(sc, legend, rect.Min.X, o.Y+3, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (g *Gauge) drawNum(sc tcell.Screen, o image.Point, n number, style tcell.Style) {
	colors := g.colorForSeries()
	if n.ok {
		style = style.Foreground(colors[0])
		printDelta(sc, n.delta, o, style)
	}

	dm, significant := NewDotMatrix(), n.val == 0
	if significant {
		style = g.dimmed
	}
	for i := range len(n.str) {
		if n.str[i] == '0' && !significant {
			g.drawDial(sc, dm.Print(int(n.str[i]-48)), o, g.dimmed)
		} else {
			significant = true
			g.drawDial(sc, dm.Print(int(n.str[i]-48)), o, style)
		}
		o.X += 3
	}
	if !n.ok {
		o.X++
		printDelta(sc, n.delta, o, style)
	}
}

func (*Gauge) drawDial(sc tcell.Screen, m Matrix, o image.Point, style tcell.Style) {
	for r := range m {
		var c int
		for c < len(m[r]) {
			dot := m[r][c]
			if dot != dots[0] {
				sc.SetContent(o.X+c, o.Y+r, dot, nil, style)
			}
			c++
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func computeDelta(d1, d2 int) delta {
	if d2 == 0 {
		return DeltaSame
	}

	d := d2 - d1
	switch {
	case d > 0:
		return DeltaMore
	case d < 0:
		return DeltaLess
	default:
		return DeltaSame
	}
}

func printDelta(sc tcell.Screen, d delta, o image.Point, s tcell.Style) {
	s = s.Dim(false)
	switch d {
	case DeltaLess:
		sc.SetContent(o.X-1, o.Y+1, '↓', nil, s)
	case DeltaMore:
		sc.SetContent(o.X-1, o.Y+1, '↑', nil, s)
	}
}

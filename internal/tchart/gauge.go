package tchart

import (
	"fmt"
	"image"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const (
	DeltaSame delta = iota
	DeltaMore
	DeltaLess

	gaugeFmt = "0%dd"
)

type delta int

// Gauge represents a gauge component.
type Gauge struct {
	*Component

	data                Metric
	deltaOk, deltaFault delta
}

// NewGauge returns a new gauge.
func NewGauge(id string) *Gauge {
	return &Gauge{
		Component: NewComponent(id),
	}
}

// IsDial returns true if chart is a dial
func (g *Gauge) IsDial() bool {
	return true
}

func (g *Gauge) Add(m Metric) {
	g.mx.Lock()
	defer g.mx.Unlock()

	g.deltaOk, g.deltaFault = computeDelta(g.data.OK, m.OK), computeDelta(g.data.Fault, m.Fault)
	g.data = m
}

func (g *Gauge) drawNum(sc tcell.Screen, ok bool, o image.Point, n int, dn delta, ns string, style tcell.Style) {
	c1, _ := g.colorForSeries()
	if ok {
		o.X -= 1
		style = style.Foreground(c1)
		printDelta(sc, dn, o, style)
		o.X += 1
	}

	dm, sig := NewDotMatrix(5, 3), n == 0
	for i := 0; i < len(ns); i++ {
		if ns[i] == '0' && !sig {
			g.drawDial(sc, dm.Print(int(ns[i]-48)), o, g.dimmed)
		} else {
			sig = true
			g.drawDial(sc, dm.Print(int(ns[i]-48)), o, style)
		}
		o.X += 5
	}
	if !ok {
		printDelta(sc, dn, o, style)
	}
}

func (g *Gauge) Draw(sc tcell.Screen) {
	g.Component.Draw(sc)

	g.mx.RLock()
	defer g.mx.RUnlock()

	rect := g.asRect()
	mid := image.Point{X: rect.Min.X + rect.Dx()/2 - 2, Y: rect.Min.Y + rect.Dy()/2 - 2}

	style := tcell.StyleDefault.Background(g.bgColor)
	style = style.Foreground(tcell.ColorYellow)
	sc.SetContent(mid.X+1, mid.Y+2, '⠔', nil, style)

	var (
		max  = g.data.MaxDigits()
		fmat = "%" + fmt.Sprintf(gaugeFmt, max)
		o    = image.Point{X: mid.X - 3, Y: mid.Y}
	)

	s1C, s2C := g.colorForSeries()
	d1, d2 := fmt.Sprintf(fmat, g.data.OK), fmt.Sprintf(fmat, g.data.Fault)
	o.X -= (len(d1) - 1) * 5
	g.drawNum(sc, true, o, g.data.OK, g.deltaOk, d1, style.Foreground(s1C).Dim(false))

	o.X = mid.X + 3
	g.drawNum(sc, false, o, g.data.Fault, g.deltaFault, d2, style.Foreground(s2C).Dim(false))

	if rect.Dx() > 0 && rect.Dy() > 0 && g.legend != "" {
		legend := g.legend
		if g.HasFocus() {
			legend = "[:aqua]" + g.legend + "[::]"
		}
		tview.Print(sc, legend, rect.Min.X, rect.Max.Y-1, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (g *Gauge) drawDial(sc tcell.Screen, m Matrix, o image.Point, style tcell.Style) {
	for r := 0; r < len(m); r++ {
		for c := 0; c < len(m[r]); c++ {
			dot := m[r][c]
			if dot == dots[0] {
				sc.SetContent(o.X+c, o.Y+r, dots[1], nil, g.dimmed)
			} else {
				sc.SetContent(o.X+c, o.Y+r, dot, nil, style)
			}
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
		sc.SetContent(o.X-1, o.Y+2, '↓', nil, s)
	case DeltaMore:
		sc.SetContent(o.X-1, o.Y+2, '↑', nil, s)
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"fmt"
	"image"

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

	gaugeFmt = "0%dd"
)

type delta int

// Gauge represents a gauge component.
type Gauge struct {
	*Component

	data             Metric
	resolution       int
	deltaOk, deltaS2 delta
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

// IsDial returns true if chart is a dial.
func (g *Gauge) IsDial() bool {
	return true
}

// Add adds a new metric.
func (g *Gauge) Add(m Metric) {
	g.mx.Lock()
	defer g.mx.Unlock()

	g.deltaOk, g.deltaS2 = computeDelta(g.data.S1, m.S1), computeDelta(g.data.S2, m.S2)
	g.data = m
}

type number struct {
	ok    bool
	val   int64
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
	style := tcell.StyleDefault.Background(g.bgColor)
	style = style.Foreground(tcell.ColorYellow)
	sc.SetContent(mid.X, mid.Y, '⠔', nil, style)

	max := g.data.MaxDigits()
	if max < g.resolution {
		max = g.resolution
	}
	var (
		fmat = "%" + fmt.Sprintf(gaugeFmt, max)
		o    = image.Point{X: mid.X, Y: mid.Y - 1}
	)

	s1C, s2C := g.colorForSeries()
	d1, d2 := fmt.Sprintf(fmat, g.data.S1), fmt.Sprintf(fmat, g.data.S2)
	o.X -= len(d1) * 3
	g.drawNum(sc, o, number{ok: true, val: g.data.S1, delta: g.deltaOk, str: d1}, style.Foreground(s1C).Dim(false))

	o.X = mid.X + 1
	g.drawNum(sc, o, number{ok: false, val: g.data.S2, delta: g.deltaS2, str: d2}, style.Foreground(s2C).Dim(false))

	if rect.Dx() > 0 && rect.Dy() > 0 && g.legend != "" {
		legend := g.legend
		if g.HasFocus() {
			legend = fmt.Sprintf("[%s:%s:]", g.focusFgColor, g.focusBgColor) + g.legend + "[::]"
		}
		tview.Print(sc, legend, rect.Min.X, o.Y+3, rect.Dx(), tview.AlignCenter, tcell.ColorWhite)
	}
}

func (g *Gauge) drawNum(sc tcell.Screen, o image.Point, n number, style tcell.Style) {
	c1, _ := g.colorForSeries()
	if n.ok {
		style = style.Foreground(c1)
		printDelta(sc, n.delta, o, style)
	}

	dm, significant := NewDotMatrix(), n.val == 0
	if significant {
		style = g.dimmed
	}
	for i := 0; i < len(n.str); i++ {
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

func computeDelta(d1, d2 int64) delta {
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
	// nolint:exhaustive
	switch d {
	case DeltaLess:
		sc.SetContent(o.X-1, o.Y+1, '↓', nil, s)
	case DeltaMore:
		sc.SetContent(o.X-1, o.Y+1, '↑', nil, s)
	}
}

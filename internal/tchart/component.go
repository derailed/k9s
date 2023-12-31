// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"image"
	"sync"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	okColor, faultColor         = tcell.ColorPaleGreen, tcell.ColorOrangeRed
	okColorName, faultColorName = "palegreen", "orangered"
)

// Component represents a graphic component.
type Component struct {
	*tview.Box

	bgColor, noColor           tcell.Color
	focusFgColor, focusBgColor string
	seriesColors               []tcell.Color
	dimmed                     tcell.Style
	id, legend                 string
	blur                       func(tcell.Key)
	mx                         sync.RWMutex
}

// NewComponent returns a new component.
func NewComponent(id string) *Component {
	return &Component{
		Box:          tview.NewBox(),
		id:           id,
		noColor:      tcell.ColorDefault,
		seriesColors: []tcell.Color{tview.Styles.PrimaryTextColor, tview.Styles.FocusColor},
		dimmed:       tcell.StyleDefault.Background(tview.Styles.PrimitiveBackgroundColor).Foreground(tcell.ColorGray).Dim(true),
	}
}

// SetFocusColorNames sets the focus color names.
func (c *Component) SetFocusColorNames(fg, bg string) {
	c.focusFgColor, c.focusBgColor = fg, bg
}

// SetBackgroundColor sets the graph bg color.
func (c *Component) SetBackgroundColor(color tcell.Color) {
	c.Box.SetBackgroundColor(color)
	c.bgColor = color
	c.dimmed = c.dimmed.Background(color)
}

// ID returns the component ID.
func (c *Component) ID() string {
	return c.id
}

// SetLegend sets the component legend.
func (c *Component) SetLegend(l string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.legend = l
}

// InputHandler returns the handler for this primitive.
func (c *Component) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// nolint:exhaustive
		switch key := event.Key(); key {
		case tcell.KeyEnter:
		case tcell.KeyBacktab, tcell.KeyTab:
			if c.blur != nil {
				c.blur(key)
			}
			setFocus(c)
		}
	})
}

// IsDial returns true if chart is a dial.
func (c *Component) IsDial() bool {
	return false
}

// SetBlurFunc sets a callback fn when component gets out of focus.
func (c *Component) SetBlurFunc(handler func(key tcell.Key)) *Component {
	c.blur = handler
	return c
}

// SetSeriesColors sets the component series colors.
func (c *Component) SetSeriesColors(cc ...tcell.Color) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.seriesColors = cc
}

// GetSeriesColorNames returns series colors by name.
func (c *Component) GetSeriesColorNames() []string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	nn := make([]string, 0, len(c.seriesColors))
	for _, color := range c.seriesColors {
		for name, co := range tcell.ColorNames {
			if co == color {
				nn = append(nn, name)
			}
		}
	}
	if len(nn) < 2 {
		nn = append(nn, okColorName, faultColorName)
	}

	return nn
}

func (c *Component) colorForSeries() (tcell.Color, tcell.Color) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	if len(c.seriesColors) == 2 {
		return c.seriesColors[0], c.seriesColors[1]
	}
	return okColor, faultColor
}

func (c *Component) asRect() image.Rectangle {
	x, y, width, height := c.GetInnerRect()
	return image.Rectangle{
		Min: image.Point{X: x, Y: y},
		Max: image.Point{X: x + width, Y: y + height},
	}
}

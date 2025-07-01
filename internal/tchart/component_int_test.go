package tchart

import (
	"image"
	"testing"

	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestComponentAsRect(t *testing.T) {
	c := NewComponent("fred")
	r := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 15, Y: 10}}

	assert.Equal(t, r, c.asRect())
}

func TestComponentColorForSeries(t *testing.T) {
	c := NewComponent("fred")
	cc := c.colorForSeries()

	assert.Len(t, cc, 3)
	assert.Equal(t, tcell.ColorGreen, cc[0])
	assert.Equal(t, tcell.ColorOrange, cc[1])
	assert.Equal(t, tcell.ColorOrangeRed, cc[2])
	assert.Equal(t, []string{"green", "orange", "orangered"}, c.GetSeriesColorNames())
}

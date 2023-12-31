// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"image"
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestComponentAsRect(t *testing.T) {
	c := NewComponent("fred")
	r := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 15, Y: 10}}

	assert.Equal(t, r, c.asRect())
}

func TestComponentColorForSeries(t *testing.T) {
	c := NewComponent("fred")
	okC, errC := c.colorForSeries()

	assert.Equal(t, tview.Styles.PrimaryTextColor, okC)
	assert.Equal(t, tview.Styles.FocusColor, errC)
	assert.Equal(t, []string{"white", "green"}, c.GetSeriesColorNames())
}

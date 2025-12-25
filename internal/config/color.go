// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	// DefaultColor represents  a default color.
	DefaultColor Color = "default"

	// TransparentColor represents the terminal bg color.
	TransparentColor Color = "-"
)

// Colors tracks multiple colors.
type Colors []Color

// Colors converts series string colors to colors.
func (c Colors) Colors() []tcell.Color {
	cc := make([]tcell.Color, 0, len(c))
	for _, color := range c {
		cc = append(cc, color.Color())
	}

	return cc
}

// Invert returns a new Colors with all colors inverted.
func (c Colors) Invert() Colors {
	inverted := make(Colors, len(c))
	for i, color := range c {
		inverted[i] = color.InvertColor()
	}
	return inverted
}

// Color represents a color.
type Color string

// NewColor returns a new color.
func NewColor(c string) Color {
	return Color(c)
}

// String returns color as string.
func (c Color) String() string {
	if c.isHex() {
		return string(c)
	}
	if c == DefaultColor {
		return "-"
	}
	col := c.Color().TrueColor().Hex()
	if col < 0 {
		return "-"
	}

	return fmt.Sprintf("#%06x", col)
}

func (c Color) isHex() bool {
	return len(c) == 7 && c[0] == '#'
}

// Color returns a view color.
func (c Color) Color() tcell.Color {
	if c == DefaultColor {
		return tcell.ColorDefault
	}

	return tcell.GetColor(string(c)).TrueColor()
}

// maxChromaForLH finds the maximum chroma at a given lightness and hue
// that stays within the sRGB gamut using binary search.
func maxChromaForLH(L, h float64) float64 {
	lo, hi := 0.0, 0.4
	for hi-lo > 0.001 {
		mid := (lo + hi) / 2
		col := colorful.OkLch(L, mid, h)
		if col.IsValid() {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// InvertColor inverts the color's lightness in Oklch space while preserving
// relative chroma (saturation relative to the maximum possible at that lightness).
// Special colors (default, transparent) are returned unchanged.
func (c Color) InvertColor() Color {
	if c == DefaultColor || c == TransparentColor || c == "" {
		return c
	}

	tc := c.Color()
	if tc == tcell.ColorDefault {
		return c
	}

	hex := tc.TrueColor().Hex()
	if hex < 0 {
		return c
	}

	col, err := colorful.Hex(fmt.Sprintf("#%06x", hex))
	if err != nil {
		return c
	}

	L, C, h := col.OkLch()

	maxC := maxChromaForLH(L, h)
	relativeChroma := 0.0
	if maxC > 0 {
		relativeChroma = C / maxC
	}

	invertedL := 1.0 - L

	maxCInverted := maxChromaForLH(invertedL, h)
	invertedC := relativeChroma * maxCInverted

	inverted := colorful.OkLch(invertedL, invertedC, h).Clamped()

	return NewColor(inverted.Hex())
}

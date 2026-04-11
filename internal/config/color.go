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
func maxChromaForLH(l, h float64) float64 {
	lo, hi := 0.0, 0.4
	for hi-lo > 0.001 {
		mid := (lo + hi) / 2
		col := colorful.OkLch(l, mid, h)
		if col.IsValid() {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// chromaPreserveFactor controls how much original chroma to preserve during
// inversion. 0.5 means we try to keep 50% of the original chroma, which
// provides a good balance between color differentiation and L inversion.
const chromaPreserveFactor = 0.5

// closestLForChroma finds the L value closest to targetL that can support
// the given chroma at the given hue. It searches toward 0.5 first (where
// gamut is typically larger), then away from 0.5 if needed.
func closestLForChroma(targetL, c, h float64) float64 {
	if maxChromaForLH(targetL, h) >= c {
		return targetL
	}

	// Search toward 0.5 first (where gamut is larger)
	if targetL < 0.5 {
		for ll := targetL; ll <= 0.5; ll += 0.01 {
			if maxChromaForLH(ll, h) >= c {
				return ll
			}
		}
		// Continue searching above 0.5 if needed
		for ll := 0.51; ll <= 0.95; ll += 0.01 {
			if maxChromaForLH(ll, h) >= c {
				return ll
			}
		}
	} else {
		for ll := targetL; ll >= 0.5; ll -= 0.01 {
			if maxChromaForLH(ll, h) >= c {
				return ll
			}
		}
		// Continue searching below 0.5 if needed
		for ll := 0.49; ll >= 0.05; ll -= 0.01 {
			if maxChromaForLH(ll, h) >= c {
				return ll
			}
		}
	}

	return targetL
}

// InvertColor inverts the color's lightness in Oklch space while preserving
// chroma (saturation). For chromatic colors, L is adjusted toward 0.5 only
// as needed to preserve a fraction of the original chroma (set by
// chromaPreserveFactor), since the sRGB gamut has less room for chroma at
// extreme lightness values.
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

	// For achromatic colors, simply invert L
	if C < 0.01 {
		return NewColor(colorful.OkLch(1.0-L, 0, h).Clamped().Hex())
	}

	// For chromatic colors, find L closest to inverted that preserves
	// at least chromaPreserveFactor of the original chroma
	targetL := 1.0 - L
	minC := C * chromaPreserveFactor
	actualL := closestLForChroma(targetL, minC, h)

	// Use as much of the original chroma as the gamut allows at actualL
	maxC := maxChromaForLH(actualL, h)
	actualC := C
	if maxC < C {
		actualC = maxC
	}

	inverted := colorful.OkLch(actualL, actualC, h).Clamped()

	return NewColor(inverted.Hex())
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"math"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	uu := map[string]struct {
		cc []string
		ee []tcell.Color
	}{
		"empty": {
			ee: []tcell.Color{},
		},
		"default": {
			cc: []string{"default"},
			ee: []tcell.Color{tcell.ColorDefault},
		},
		"multi": {
			cc: []string{
				"default",
				"transparent",
				"blue",
				"green",
			},
			ee: []tcell.Color{
				tcell.ColorDefault,
				tcell.ColorDefault,
				tcell.ColorBlue.TrueColor(),
				tcell.ColorGreen.TrueColor(),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cc := make(config.Colors, 0, len(u.cc))
			for _, c := range u.cc {
				cc = append(cc, config.NewColor(c))
			}
			assert.Equal(t, u.ee, cc.Colors())
		})
	}
}

func TestColorString(t *testing.T) {
	uu := map[string]struct {
		c string
		e string
	}{
		"empty": {
			e: "-",
		},
		"default": {
			c: "default",
			e: "-",
		},
		"transparent": {
			c: "-",
			e: "-",
		},
		"blue": {
			c: "blue",
			e: "#0000ff",
		},
		"lightgray": {
			c: "lightgray",
			e: "#d3d3d3",
		},
		"hex": {
			c: "#00ff00",
			e: "#00ff00",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := config.NewColor(u.c)
			assert.Equal(t, u.e, c.String())
		})
	}
}

func TestColorToColor(t *testing.T) {
	uu := map[string]struct {
		c string
		e tcell.Color
	}{
		"default": {
			c: "default",
			e: tcell.ColorDefault,
		},
		"transparent": {
			c: "-",
			e: tcell.ColorDefault,
		},
		"aqua": {
			c: "aqua",
			e: tcell.ColorAqua.TrueColor(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := config.NewColor(u.c)
			assert.Equal(t, u.e, c.Color())
		})
	}
}

// getOkch returns c, h for a hex color string.
func getOkch(hex string) (c, h float64) {
	col, err := colorful.Hex(hex)
	if err != nil {
		return 0, 0
	}
	_, c, h = col.OkLch()
	return c, h
}

// huesEqual checks if two hues are equal within tolerance, handling wraparound.
func huesEqual(h1, h2, tolerance float64) bool {
	diff := math.Abs(h1 - h2)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff < tolerance
}

func TestInvertColor(t *testing.T) {
	uu := map[string]struct {
		c      string
		expect string
	}{
		"default": {
			c:      "default",
			expect: "default",
		},
		"transparent": {
			c:      "-",
			expect: "-",
		},
		"empty": {
			c:      "",
			expect: "",
		},
		"black_to_white": {
			c:      "#000000",
			expect: "#ffffff",
		},
		"white_to_black": {
			c:      "#ffffff",
			expect: "#000000",
		},
		"red_to_dark": {
			// L=0.628, C=0.258, h=29.2
			c:      "#ff0000",
			expect: "#7e0000",
		},
		"blue_to_light": {
			// L=0.452, C=0.313, h=264.1 -> L adjusted to 0.55 to preserve chroma
			c:      "#0000ff",
			expect: "#1f5bff",
		},
		"green_to_dark": {
			// L=0.866, C=0.295, h=142.5 -> L adjusted to 0.44 to preserve chroma
			c:      "#00ff00",
			expect: "#006600",
		},
		"yellow_to_dark": {
			// L=0.968, C=0.211, h=109.8 -> L adjusted to 0.49 to preserve chroma
			c:      "#ffff00",
			expect: "#656501",
		},
		"cyan_to_dark": {
			// L=0.905, C=0.155, h=194.8 -> L adjusted to 0.46 to preserve chroma
			c:      "#00ffff",
			expect: "#016464",
		},
		"dark_gray_to_light": {
			c:      "#333333",
			expect: "#989898",
		},
		"light_gray_to_dark": {
			c:      "#cccccc",
			expect: "#0c0c0c",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := config.NewColor(u.c)
			inverted := c.InvertColor()
			assert.Equal(t, u.expect, string(inverted))
		})
	}
}

func TestInvertColorPreservesHue(t *testing.T) {
	// Verify that hue is preserved during inversion for chromatic colors.
	// Note: Hue preservation depends on the inverted color having sufficient chroma
	// and not being clamped by go-colorful's Clamped() method. Colors near the
	// gamut boundary may have hue shifts after clamping.
	uu := map[string]struct {
		c string
		h float64 // expected hue
	}{
		"red": {
			// L=0.628, C=0.258, h=29.2 -> inverted to L=0.372 with good chroma
			c: "#ff0000",
			h: 29.2,
		},
		"blue": {
			// L=0.452, C=0.313, h=264.1 -> inverted to L=0.548 with good chroma
			c: "#0000ff",
			h: 264.1,
		},
		"purple": {
			// L=0.420, C=0.161, h=328.4 -> mid-lightness, stable hue
			c: "#800080",
			h: 328.4,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			original := config.NewColor(u.c)
			inverted := original.InvertColor()

			_, hOrig := getOkch(u.c)
			cInv, hInv := getOkch(string(inverted))

			// Only check hue if inverted color has meaningful chroma (C > 0.05)
			// Below this threshold, sRGB quantization causes hue instability
			if cInv > 0.05 {
				assert.True(t, huesEqual(hOrig, hInv, 1.0),
					"hue should be preserved: original h=%.1f, inverted h=%.1f", hOrig, hInv)
			}
		})
	}
}

func TestInvertGrayRoundTrip(t *testing.T) {
	// Achromatic colors (grays) round-trip perfectly because they have no chroma
	// to lose during gamut-constrained scaling.
	colors := []string{
		"#000000",
		"#ffffff",
		"#808080",
		"#333333",
		"#cccccc",
		"#555555",
		"#636363", // L=0.5 in Oklch
	}

	for _, c := range colors {
		t.Run(c, func(t *testing.T) {
			original := config.NewColor(c)
			inverted := original.InvertColor()
			reinverted := inverted.InvertColor()

			assert.Equal(t, original.String(), string(reinverted),
				"double inversion should return to original for achromatic colors")
		})
	}
}

func TestInvertColorSelfInverting(t *testing.T) {
	// Colors with L=0.5 in Oklch invert to themselves.
	// For achromatic grays, L=0.5 corresponds to approximately #636363 in sRGB.
	selfInverting := []string{
		"#636363",
	}

	for _, c := range selfInverting {
		t.Run(c, func(t *testing.T) {
			original := config.NewColor(c)
			inverted := original.InvertColor()

			assert.Equal(t, original.String(), string(inverted),
				"color with L=0.5 should invert to itself")
		})
	}
}

func TestInvertColorOutOfGamut(t *testing.T) {
	// These highly saturated colors would produce out-of-gamut results if we
	// simply inverted L without adjustment. The chroma-preserving approach
	// finds an L closer to 0.5 where sufficient chroma is available.
	//
	// For colors with very high L (yellow, cyan), the ideal inverted L would
	// be very low where max chroma is tiny. Instead, L is adjusted toward 0.5
	// to preserve chromaPreserveFactor (0.5) of the original chroma.
	uu := map[string]struct {
		c      string
		expect string
	}{
		"saturated_yellow": {
			// L=0.968, C=0.211 -> L adjusted to 0.49 to preserve 50% chroma
			c:      "#ffff00",
			expect: "#656501",
		},
		"saturated_cyan": {
			// L=0.905, C=0.155 -> L adjusted to 0.46 to preserve 50% chroma
			c:      "#00ffff",
			expect: "#016464",
		},
		"saturated_blue": {
			// L=0.452, C=0.313 -> L adjusted to 0.55 to preserve chroma
			c:      "#0000ff",
			expect: "#1f5bff",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			original := config.NewColor(u.c)
			inverted := original.InvertColor()

			// Verify the inverted color matches expected
			invertedStr := string(inverted)
			assert.Equal(t, u.expect, invertedStr)

			// Verify the inverted color is valid hex
			assert.Regexp(t, `^#[0-9a-f]{6}$`, invertedStr,
				"inverted color should be valid hex")

			// Verify it differs from original (these are not L=0.5 colors)
			assert.NotEqual(t, original.String(), invertedStr,
				"saturated color should not invert to itself")

			// Only check hue preservation for colors with meaningful inverted chroma (C > 0.05)
			// Below this threshold, sRGB quantization causes hue instability
			cInv, hInv := getOkch(invertedStr)
			if cInv > 0.05 {
				_, hOrig := getOkch(u.c)
				assert.True(t, huesEqual(hOrig, hInv, 1.0),
					"hue should be preserved: original h=%.1f, inverted h=%.1f", hOrig, hInv)
			}
		})
	}
}

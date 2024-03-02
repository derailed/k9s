// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog/log"
)

const (
	// DefaultColorName indicator to keep term colors.
	DefaultColorName = "default"

	// SearchFmt represents a filter view title.
	SearchFmt = "<[filter:bg:r]/%s[fg:bg:-]> "

	// NSTitleFmt represents a namespaced view title.
	NSTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "

	// TitleFmt represents a standard view title.
	TitleFmt = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "

	descIndicator = "↓"
	ascIndicator  = "↑"

	// FullFmat specifies a namespaced dump file name.
	FullFmat = "%s-%s-%d.csv"

	// NoNSFmat specifies a cluster wide dump file name.
	NoNSFmat = "%s-%d.csv"
)

func mustExtractStyles(ctx context.Context) *config.Styles {
	styles, ok := ctx.Value(internal.KeyStyles).(*config.Styles)
	if !ok {
		log.Fatal().Msg("Expecting valid styles")
	}
	return styles
}

// TrimCell removes superfluous padding.
func TrimCell(tv *SelectTable, row, col int) string {
	c := tv.GetCell(row, col)
	if c == nil {
		log.Error().Err(fmt.Errorf("No cell at location [%d:%d]", row, col)).Msg("Trim cell failed!")
		return ""
	}
	return strings.TrimSpace(c.Text)
}

// TrimLabelSelector extracts label query.
func TrimLabelSelector(s string) string {
	if strings.Index(s, "-l") == 0 {
		return strings.TrimSpace(s[2:])
	}

	return s
}

// SkinTitle decorates a title.
func SkinTitle(fmat string, style config.Frame) string {
	bgColor := style.Title.BgColor
	if bgColor == config.DefaultColor {
		bgColor = config.TransparentColor
	}
	fmat = strings.Replace(fmat, "[fg:bg", "["+style.Title.FgColor.String()+":"+bgColor.String(), -1)
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor.String(), 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor.String(), 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor.String(), 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor.String(), 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+bgColor.String()+":", -1)

	return fmat
}

func sortIndicator(sort, asc bool, style config.Table, name string) string {
	if !sort {
		return name
	}

	order := descIndicator
	if asc {
		order = ascIndicator
	}
	return fmt.Sprintf("%s[%s::b]%s[::]", name, style.Header.SorterColor, order)
}

func formatCell(field string, padding int) string {
	if IsASCII(field) {
		return Pad(field, padding)
	}

	return field
}

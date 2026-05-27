// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	// DefaultColorName indicator to keep term colors.
	DefaultColorName = "default"

	// SearchFmt represents a filter view title.
	SearchFmt = "<[filter:bg:r]/%s[fg:bg:-]> "

	// NSTitleFmt represents a namespaced view title.
	NSTitleFmt = " [fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "

	// TitleFmt represents a standard view title.
	TitleFmt = " [fg:bg:b]%s[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "

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
		slog.Error("Expecting valid styles. Exiting!")
		os.Exit(1)
	}
	return styles
}

// TrimCell removes superfluous padding.
func TrimCell(tv *SelectTable, row, col int) string {
	c := tv.GetCell(row, col)
	if c == nil {
		slog.Error("Trim cell failed", slogs.Error, fmt.Errorf("no cell at [%d:%d]", row, col))
		return ""
	}
	return strings.TrimSpace(c.Text)
}

// ExtractLabelSelector extracts label query.
func ExtractLabelSelector(s string) (labels.Selector, error) {
	selStr := s
	if strings.Index(s, "-l") == 0 {
		selStr = strings.TrimSpace(s[2:])
	}

	return labels.Parse(selStr)
}

// SkinTitle decorates a title.
func SkinTitle(fmat string, style *config.Frame) string {
	bgColor := style.Title.BgColor
	if bgColor == config.DefaultColor {
		bgColor = config.TransparentColor
	}
	fmat = strings.ReplaceAll(fmat, "[fg:bg", "["+style.Title.FgColor.String()+":"+bgColor.String())
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor.String(), 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor.String(), 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor.String(), 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor.String(), 1)
	fmat = strings.ReplaceAll(fmat, ":bg:", ":"+bgColor.String()+":")

	return fmat
}

func columnIndicator(sortRank int, selected, asc bool, style *config.Table, name string) string {
	var displayName string
	if selected {
		displayName = fmt.Sprintf("[%s::]%s[::]", style.Header.SelectedSortColumnColor, name)
	} else {
		displayName = fmt.Sprintf("[%s::]%s[::]", style.Header.FgColor, name)
	}

	suffix := ""
	if sortRank > 0 {
		order := descIndicator
		if asc {
			order = ascIndicator
		}
		var color config.Color
		switch sortRank {
		case 1:
			color = style.Header.SorterColor
		case 2:
			color = style.Header.SecondarySorterColor
		default:
			color = style.Header.TertiarySorterColor
		}
		if sortRank == 1 {
			suffix = fmt.Sprintf("[%s::b]%s[::]", color, order)
		} else {
			suffix = fmt.Sprintf("[%s::d]%s[::]", color, order)
		}
	}

	return displayName + suffix
}

func formatCell(field string, padding int) string {
	if IsASCII(field) {
		return Pad(field, padding)
	}

	return field
}

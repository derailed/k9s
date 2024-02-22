// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
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

var (
	// LabelRx identifies a label query.
	LabelRx = regexp.MustCompile(`\A\-l`)
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

// IsLabelSelector checks if query is a label query.
func IsLabelSelector(s string) bool {
	if LabelRx.MatchString(s) {
		return true
	}

	return !strings.Contains(s, " ") && cmd.ToLabels(s) != nil
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

func filterToast(data *render.TableData) *render.TableData {
	validCol := data.Header.IndexOf("VALID", true)
	if validCol == -1 {
		return data
	}

	toast := render.TableData{
		Header:    data.Header,
		RowEvents: render.NewRowEvents(data.RowEvents.Len()),
		Namespace: data.Namespace,
	}
	data.RowEvents.Range(func(_ int, re render.RowEvent) bool {
		if re.Row.Fields[validCol] != "" {
			toast.RowEvents.Add(re)
		}
		return true
	})

	return &toast
}

func rxFilter(q string, inverse bool, data *render.TableData) (*render.TableData, error) {
	if inverse {
		q = q[1:]
	}
	rx, err := regexp.Compile(`(?i)(` + q + `)`)
	if err != nil {
		return data, fmt.Errorf("%w -- %s", err, q)
	}

	filtered := render.TableData{
		Header:    data.Header,
		RowEvents: render.NewRowEvents(data.RowEvents.Len()),
		Namespace: data.Namespace,
	}
	ageIndex := data.Header.IndexOf("AGE", true)

	const spacer = " "
	data.RowEvents.Range(func(_ int, re render.RowEvent) bool {
		ff := re.Row.Fields
		if ageIndex >= 0 && ageIndex+1 <= len(ff) {
			ff = append(ff[0:ageIndex], ff[ageIndex+1:]...)
		}
		fields := strings.Join(ff, spacer)
		if (inverse && !rx.MatchString(fields)) ||
			((!inverse) && rx.MatchString(fields)) {
			filtered.RowEvents.Add(re)
		}

		return true
	})

	return &filtered, nil
}

func fuzzyFilter(q string, data *render.TableData) *render.TableData {
	q = strings.TrimSpace(q)
	ss := make([]string, 0, data.RowEvents.Len())
	data.RowEvents.Range(func(_ int, re render.RowEvent) bool {
		ss = append(ss, re.Row.ID)
		return true
	})

	filtered := render.TableData{
		Header:    data.Header,
		RowEvents: render.NewRowEvents(data.RowEvents.Len()),
		Namespace: data.Namespace,
	}
	mm := fuzzy.Find(q, ss)
	for _, m := range mm {
		re, ok := data.RowEvents.At(m.Index)
		if !ok {
			log.Error().Msgf("unable to find event for index in fuzzfilter: %d", m.Index)
			continue
		}
		filtered.RowEvents.Add(re)
	}

	return &filtered
}

package ui

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

const (
	// SearchFmt represents a filter view title.
	SearchFmt = "<[filter:bg:r]/%s[fg:bg:-]> "

	nsTitleFmt    = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
	titleFmt      = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
	descIndicator = "↓"
	ascIndicator  = "↑"

	// FullFmat specifies a namespaced dump file name.
	FullFmat = "%s-%s-%d.csv"

	// NoNSFmat specifies a cluster wide dump file name.
	NoNSFmat = "%s-%d.csv"
)

var (
	// LabelCmd identifies a label query
	LabelCmd = regexp.MustCompile(`\A\-l`)

	fuzzyCmd = regexp.MustCompile(`\A\-f`)
)

func mustExtractSyles(ctx context.Context) *config.Styles {
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
	if s == "" {
		return false
	}
	return LabelCmd.MatchString(s)
}

// IsFuzztySelector checks if query is fuzzy.
func isFuzzySelector(s string) bool {
	if s == "" {
		return false
	}
	return fuzzyCmd.MatchString(s)
}

// TrimLabelSelector extracts label query.
func TrimLabelSelector(s string) string {
	return strings.TrimSpace(s[2:])
}

// SkinTitle decorates a title.
func SkinTitle(fmat string, style config.Frame) string {
	fmat = strings.Replace(fmat, "[fg:bg", "["+style.Title.FgColor+":"+style.Title.BgColor, -1)
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor, 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor, 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor, 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+style.Title.BgColor+":", -1)

	return fmat
}

func sortIndicator(col SortColumn, style config.Table, index int, name string) string {
	if col.index != index {
		return name
	}

	order := descIndicator
	if col.asc {
		order = ascIndicator
	}
	return fmt.Sprintf("%s[%s::]%s[::]", name, style.Header.SorterColor, order)
}

func formatCell(field string, padding int) string {
	if IsASCII(field) {
		return Pad(field, padding)
	}

	return field
}

func rxFilter(q string, data render.TableData) (render.TableData, error) {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return data, err
	}

	filtered := render.TableData{
		Header:    data.Header,
		RowEvents: make(render.RowEvents, 0, len(data.RowEvents)),
		Namespace: data.Namespace,
	}
	for _, re := range data.RowEvents {
		f := strings.Join(re.Row.Fields, " ")
		if rx.MatchString(f) {
			filtered.RowEvents = append(filtered.RowEvents, re)
		}
	}

	return filtered, nil
}

func fuzzyFilter(q string, index int, data render.TableData) render.TableData {
	var ss []string
	for _, re := range data.RowEvents {
		ss = append(ss, re.Row.Fields[index])
	}

	filtered := render.TableData{
		Header:    data.Header,
		RowEvents: make(render.RowEvents, 0, len(data.RowEvents)),
		Namespace: data.Namespace,
	}
	mm := fuzzy.Find(q, ss)
	for _, m := range mm {
		filtered.RowEvents = append(filtered.RowEvents, data.RowEvents[m.Index])
	}

	return filtered
}

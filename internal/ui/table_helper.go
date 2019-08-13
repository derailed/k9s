package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
)

const (
	titleFmt          = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%d[fg:bg:-]] "
	searchFmt         = "<[filter:bg:r]/%s[fg:bg:-]> "
	nsTitleFmt        = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
	labelSelIndicator = "-l"
	descIndicator     = "↓"
	ascIndicator      = "↑"
	fullFmat          = "%s-%s-%d.csv"
	noNSFmat          = "%s-%d.csv"
)

var (
	cpuRX    = regexp.MustCompile(`\A.{0,1}CPU`)
	memRX    = regexp.MustCompile(`\A.{0,1}MEM`)
	labelCmd = regexp.MustCompile(`\A\-l`)
	fuzzyCmd = regexp.MustCompile(`\A\-f`)
)

type cleanseFn func(string) string

// TrimCell removes superfluous padding.
func TrimCell(tv *Table, row, col int) string {
	c := tv.GetCell(row, col)
	if c == nil {
		log.Error().Err(fmt.Errorf("No cell at location [%d:%d]", row, col)).Msg("Trim cell failed!")
		return ""
	}
	return strings.TrimSpace(c.Text)
}

func isLabelSelector(s string) bool {
	if s == "" {
		return false
	}
	return labelCmd.MatchString(s)
}

func isFuzzySelector(s string) bool {
	if s == "" {
		return false
	}
	return fuzzyCmd.MatchString(s)
}

func trimLabelSelector(s string) string {
	return strings.TrimSpace(s[2:])
}

func skinTitle(fmat string, style config.Frame) string {
	fmat = strings.Replace(fmat, "[fg:bg", "["+style.Title.FgColor+":"+style.Title.BgColor, -1)
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor, 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor, 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor, 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+style.Title.BgColor+":", -1)

	return fmat
}

func sortRows(evts resource.RowEvents, sortFn SortFn, sortCol SortColumn, keys []string) {
	rows := make(resource.Rows, 0, len(evts))
	for k, r := range evts {
		rows = append(rows, append(r.Fields, k))
	}
	sortFn(rows, sortCol)

	for i, r := range rows {
		keys[i] = r[len(r)-1]
	}
}

func defaultSort(rows resource.Rows, sortCol SortColumn) {
	t := RowSorter{rows: rows, index: sortCol.index, asc: sortCol.asc}
	sort.Sort(t)
}

func sortAllRows(col SortColumn, rows resource.RowEvents, sortFn SortFn) (resource.Row, map[string]resource.Row) {
	keys := make([]string, len(rows))
	sortRows(rows, sortFn, col, keys)

	sec := make(map[string]resource.Row, len(rows))
	for _, k := range keys {
		grp := rows[k].Fields[col.index]
		sec[grp] = append(sec[grp], k)
	}

	// Performs secondary to sort by name for each groups.
	prim := make(resource.Row, 0, len(sec))
	for k, v := range sec {
		sort.Strings(v)
		prim = append(prim, k)
	}
	sort.Sort(GroupSorter{prim, col.asc})

	return prim, sec
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

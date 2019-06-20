package views

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
)

var labelCmd = regexp.MustCompile(`\A\-l`)

const (
	descIndicator = "↓"
	ascIndicator  = "↑"
)

func isLabelSelector(s string) bool {
	if s == "" {
		return false
	}
	return labelCmd.MatchString(s)
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

func sortRows(evts resource.RowEvents, sortFn sortFn, sortCol sortColumn, keys []string) {
	rows := make(resource.Rows, 0, len(evts))
	for k, r := range evts {
		rows = append(rows, append(r.Fields, k))
	}
	sortFn(rows, sortCol)

	for i, r := range rows {
		keys[i] = r[len(r)-1]
	}
}

func defaultSort(rows resource.Rows, sortCol sortColumn) {
	t := rowSorter{rows: rows, index: sortCol.index, asc: sortCol.asc}
	sort.Sort(t)
}

func sortAllRows(col sortColumn, rows resource.RowEvents, sortFn sortFn) (resource.Row, map[string]resource.Row) {
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
	sort.Sort(groupSorter{prim, col.asc})

	return prim, sec
}

func sortIndicator(col sortColumn, style config.Table, index int, name string) string {
	if col.index != index {
		return name
	}

	order := descIndicator
	if col.asc {
		order = ascIndicator
	}
	return fmt.Sprintf("%s[%s::]%s[::]", name, style.Header.SorterColor, order)
}

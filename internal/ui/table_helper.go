package ui

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	cpuRX = regexp.MustCompile(`\A.{0,1}CPU`)
	memRX = regexp.MustCompile(`\A.{0,1}MEM`)

	// LabelCmd identifies a label query
	LabelCmd = regexp.MustCompile(`\A\-l`)

	fuzzyCmd = regexp.MustCompile(`\A\-f`)
)

func mustExtractSyles(ctx context.Context) *config.Styles {
	styles, ok := ctx.Value(KeyStyles).(*config.Styles)
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

func SkinTitle(fmat string, style config.Frame) string {
	fmat = strings.Replace(fmat, "[fg:bg", "["+style.Title.FgColor+":"+style.Title.BgColor, -1)
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor, 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor, 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor, 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+style.Title.BgColor+":", -1)

	return fmat
}

// BOZO!!
// func sortRows(evts resource.RowEvents, sortFn SortFn, sortCol SortColumn, keys []string) {
// 	rows := make(resource.Rows, 0, len(evts))
// 	for k, r := range evts {
// 		rows = append(rows, append(r.Fields, k))
// 	}
// 	sortFn(rows, sortCol)

// 	for i, r := range rows {
// 		keys[i] = r[len(r)-1]
// 	}
// }

// func defaultSort(rows resource.Rows, sortCol SortColumn) {
// 	t := RowSorter{rows: rows, index: sortCol.index, asc: sortCol.asc}
// 	sort.Sort(t)
// }

// BOZO!!
// func sortAllRows(col SortColumn, rows resource.RowEvents, sortFn SortFn) (resource.Row, map[string]resource.Row) {
// 	keys := make([]string, len(rows))
// 	sortRows(rows, sortFn, col, keys)

// 	sec := make(map[string]resource.Row, len(rows))
// 	for _, k := range keys {
// 		grp := rows[k].Fields[col.index]
// 		sec[grp] = append(sec[grp], k)
// 	}

// 	// Performs secondary to sort by name for each groups.
// 	prim := make(resource.Row, 0, len(sec))
// 	for k, v := range sec {
// 		sort.Strings(v)
// 		prim = append(prim, k)
// 	}
// 	sort.Sort(GroupSorter{prim, col.asc})

// 	return prim, sec
// }

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
	var ss, kk []string
	for _, re := range data.RowEvents {
		ss = append(ss, re.Row.Fields[index])
		kk = append(kk, re.Row.ID)
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

// UpdateTitle refreshes the table title.
func styleTitle(rc int, ns, base, path, buff string, styles *config.Styles) string {
	if rc > 0 {
		rc--
	}

	if ns == render.AllNamespaces {
		ns = render.NamespaceAll
	}
	info := ns
	if path != "" {
		info = path
		cns, n := render.Namespaced(path)
		if cns == render.ClusterWide {
			info = n
		}
	}

	var title string
	if info == "" || info == render.ClusterWide {
		title = SkinTitle(fmt.Sprintf(titleFmt, base, rc), styles.Frame())
	} else {
		title = SkinTitle(fmt.Sprintf(nsTitleFmt, base, info, rc), styles.Frame())
	}
	if buff == "" {
		return title
	}

	if IsLabelSelector(buff) {
		buff = TrimLabelSelector(buff)
	}
	return title + SkinTitle(fmt.Sprintf(SearchFmt, buff), styles.Frame())
}

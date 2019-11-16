package ui

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
	"k8s.io/apimachinery/pkg/util/duration"
)

const (
	// SearchFmt represents a filter view title.
	SearchFmt = "<[filter:bg:r]/%s[fg:bg:-]> "

	titleFmt      = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%d[fg:bg:-]] "
	nsTitleFmt    = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
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

func formatCell(numerical bool, header, field string, padding int) (string, int) {
	if header == "AGE" {
		dur, err := time.ParseDuration(field)
		if err == nil {
			field = duration.HumanDuration(dur)
		}
	}

	if numerical || cpuRX.MatchString(header) || memRX.MatchString(header) {
		return field, tview.AlignRight
	}

	align := tview.AlignLeft
	if IsASCII(field) {
		return Pad(field, padding), align
	}

	return field, align
}

func rxFilter(q string, data resource.TableData) (resource.TableData, error) {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return data, err
	}

	filtered := resource.TableData{
		Header:    data.Header,
		Rows:      resource.RowEvents{},
		Namespace: data.Namespace,
	}
	for k, row := range data.Rows {
		f := strings.Join(row.Fields, " ")
		if rx.MatchString(f) {
			filtered.Rows[k] = row
		}
	}

	return filtered, nil
}

func fuzzyFilter(q string, index int, data resource.TableData) resource.TableData {
	var ss, kk []string
	for k, row := range data.Rows {
		ss = append(ss, row.Fields[index])
		kk = append(kk, k)
	}

	filtered := resource.TableData{
		Header:    data.Header,
		Rows:      resource.RowEvents{},
		Namespace: data.Namespace,
	}
	mm := fuzzy.Find(q, ss)
	for _, m := range mm {
		filtered.Rows[kk[m.Index]] = data.Rows[kk[m.Index]]
	}

	return filtered
}

// UpdateTitle refreshes the table title.
func styleTitle(rc int, ns, base, buff string, styles *config.Styles) string {
	var title string

	if rc > 0 {
		rc--
	}
	switch ns {
	case resource.NotNamespaced, "*":
		title = SkinTitle(fmt.Sprintf(titleFmt, base, rc), styles.Frame())
	default:
		if ns == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = SkinTitle(fmt.Sprintf(nsTitleFmt, base, ns, rc), styles.Frame())
	}

	if buff != "" {
		if IsLabelSelector(buff) {
			buff = TrimLabelSelector(buff)
		}
		title += SkinTitle(fmt.Sprintf(SearchFmt, buff), styles.Frame())
	}

	return title
}

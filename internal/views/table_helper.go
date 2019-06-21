package views

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
)

const (
	titleFmt          = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%d[fg:bg:-]] "
	searchFmt         = "<[filter:bg:b]/%s[fg:bg:]> "
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
)

type cleanseFn func(string) string

func trimCellRelative(tv *tableView, row, col int) string {
	return trimCell(tv, row, tv.nameColIndex()+col)
}

func trimCell(tv *tableView, row, col int) string {
	c := tv.GetCell(row, col)
	if c == nil {
		log.Error().Err(fmt.Errorf("No cell at location [%d:%d]", row, col)).Msg("Trim cell failed!")
		return ""
	}
	return strings.TrimSpace(c.Text)
}

func saveTable(cluster, name string, data resource.TableData) (string, error) {
	dir := filepath.Join(config.K9sDumpDir, cluster)
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	ns, now := data.Namespace, time.Now().UnixNano()
	if ns == resource.AllNamespaces {
		ns = resource.AllNamespace
	}
	fName := fmt.Sprintf(fullFmat, name, ns, now)
	if ns == resource.NotNamespaced {
		fName = fmt.Sprintf(noNSFmat, name, now)
	}

	path := filepath.Join(dir, fName)
	mod := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(path, mod, 0644)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return "", err
	}

	w := csv.NewWriter(file)
	w.Write(data.Header)
	for _, r := range data.Rows {
		w.Write(r.Fields)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}

	return path, nil
}

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

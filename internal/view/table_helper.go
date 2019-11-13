package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
)

func trimCellRelative(t *Table, row, col int) string {
	return ui.TrimCell(t.Table, row, t.NameColIndex()+col)
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
	fName := fmt.Sprintf(ui.FullFmat, name, ns, now)
	if ns == resource.NotNamespaced {
		fName = fmt.Sprintf(ui.NoNSFmat, name, now)
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

func skinTitle(fmat string, style config.Frame) string {
	fmat = strings.Replace(fmat, "[fg:bg", "["+style.Title.FgColor+":"+style.Title.BgColor, -1)
	fmat = strings.Replace(fmat, "[hilite", "["+style.Title.HighlightColor, 1)
	fmat = strings.Replace(fmat, "[key", "["+style.Menu.NumKeyColor, 1)
	fmat = strings.Replace(fmat, "[filter", "["+style.Title.FilterColor, 1)
	fmat = strings.Replace(fmat, "[count", "["+style.Title.CounterColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+style.Title.BgColor+":", -1)

	return fmat
}

func sortRows(evts resource.RowEvents, sortFn ui.SortFn, sortCol ui.SortColumn, keys []string) {
	rows := make(resource.Rows, 0, len(evts))
	for k, r := range evts {
		rows = append(rows, append(r.Fields, k))
	}
	sortFn(rows, sortCol)

	for i, r := range rows {
		keys[i] = r[len(r)-1]
	}
}

// func defaultSort(rows resource.Rows, sortCol ui.SortColumn) {
// 	t := rowSorter{rows: rows, index: sortCol.index, asc: sortCol.asc}
// 	sort.Sort(t)
// }

// func sortAllRows(col ui.SortColumn, rows resource.RowEvents, sortFn ui.SortFn) (resource.Row, map[string]resource.Row) {
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
// 	sort.Sort(groupSorter{prim, col.asc})

// 	return prim, sec
// }

// func sortIndicator(col ui.SortColumn, style config.Table, index int, name string) string {
// 	if col.index != index {
// 		return name
// 	}

// 	order := descIndicator
// 	if col.asc {
// 		order = ascIndicator
// 	}
// 	return fmt.Sprintf("%s[%s::]%s[::]", name, style.Header.SorterColor, order)
// }

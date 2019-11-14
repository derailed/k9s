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
	if err := w.Write(data.Header); err != nil {
		return "", err
	}
	for _, r := range data.Rows {
		if err := w.Write(r.Fields); err != nil {
			return "", err
		}
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

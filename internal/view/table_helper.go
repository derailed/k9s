package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

func trimCellRelative(t *Table, row, col int) string {
	return ui.TrimCell(t.SelectTable, row, t.NameColIndex()+col)
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
	file, err := os.OpenFile(path, mod, 0600)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("Closing file")
		}
	}()

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

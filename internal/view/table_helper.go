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
	"github.com/rs/zerolog/log"
)

func trimCellRelative(t *Table, row, col int) string {
	return ui.TrimCell(t.SelectTable, row, t.NameColIndex()+col)
}

func computeFilename(cluster, ns, title, path string) (string, error) {
	now := time.Now().UnixNano()

	dir := filepath.Join(config.K9sDumpDir, cluster)
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	name := title + "-" + strings.Replace(path, "/", "-", -1)
	if path == "" {
		name = title
	}

	var fName string
	if ns == resource.NotNamespaced {
		fName = fmt.Sprintf(ui.NoNSFmat, name, now)
	} else {
		fName = fmt.Sprintf(ui.FullFmat, name, ns, now)
	}

	return strings.ToLower(filepath.Join(dir, fName)), nil
}

func saveTable(cluster, title, path string, data resource.TableData) (string, error) {
	ns := data.Namespace
	if ns == resource.AllNamespaces {
		ns = resource.AllNamespace
	}

	fPath, err := computeFilename(cluster, ns, title, path)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Saving Table to %s", fPath)

	mod := os.O_CREATE | os.O_WRONLY
	out, err := os.OpenFile(fPath, mod, 0600)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Error().Err(err).Msg("Closing file")
		}
	}()

	w := csv.NewWriter(out)
	// BOZO!! Method on header
	headers := make([]string, len(data.Header))
	for i, h := range data.Header {
		headers[i] = h.Name
	}
	if err := w.Write([]string(headers)); err != nil {
		return "", err
	}

	for _, re := range data.RowEvents {
		if err := w.Write(re.Row.Fields); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}

	return path, nil
}

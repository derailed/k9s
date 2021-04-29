package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

func computeFilename(cluster, ns, title, path string) (string, error) {
	now := time.Now().UnixNano()

	dir := filepath.Join(config.K9sDumpDir, dao.SanitizeFilename(cluster))
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	name := title + "-" + dao.SanitizeFilename(path)
	if path == "" {
		name = title
	}

	var fName string
	if ns == client.ClusterScope {
		fName = fmt.Sprintf(ui.NoNSFmat, name, now)
	} else {
		fName = fmt.Sprintf(ui.FullFmat, name, ns, now)
	}

	return strings.ToLower(filepath.Join(dir, fName)), nil
}

func saveTable(cluster, title, path string, data render.TableData) (string, error) {
	ns := data.Namespace
	if client.IsClusterWide(ns) {
		ns = client.NamespaceAll
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
	if err := w.Write(data.Header.Columns(true)); err != nil {
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

	return fPath, nil
}

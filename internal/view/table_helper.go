// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

func computeFilename(dumpPath, ns, title, path string) (string, error) {
	now := time.Now().UnixNano()

	dir := filepath.Join(dumpPath)
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	name := title + "-" + data.SanitizeFileName(path)
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

func saveTable(dir, title, path string, data *model1.TableData) (string, error) {
	ns := data.GetNamespace()
	if client.IsClusterWide(ns) {
		ns = client.NamespaceAll
	}

	fPath, err := computeFilename(dir, ns, title, path)
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
	_ = w.Write(data.ColumnNames(true))

	data.RowsRange(func(_ int, re model1.RowEvent) bool {
		_ = w.Write(re.Row.Fields)
		return true
	})
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}

	return fPath, nil
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// Benchmark represents a service benchmark results view.
type Benchmark struct {
	ResourceViewer
}

// NewBenchmark returns a new viewer.
func NewBenchmark(gvr client.GVR) ResourceViewer {
	b := Benchmark{
		ResourceViewer: NewBrowser(gvr),
	}
	b.GetTable().SetBorderFocusColor(tcell.ColorSeaGreen)
	b.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorSeaGreen).Attributes(tcell.AttrNone))
	b.GetTable().SetSortCol(ageCol, true)
	b.SetContextFn(b.benchContext)
	b.GetTable().SetEnterFn(b.viewBench)

	return &b
}

func (b *Benchmark) benchContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyDir, benchDir(b.App().Config))
}

func (b *Benchmark) viewBench(app *App, model ui.Tabular, gvr client.GVR, path string) {
	data, err := readBenchFile(app.Config, b.benchFile())
	if err != nil {
		app.Flash().Errf("Unable to load bench file %s", err)
		return
	}

	details := NewDetails(b.App(), "Results", fileToSubject(path), contentYAML, false).Update(data)
	if err := app.inject(details, false); err != nil {
		app.Flash().Err(err)
	}
}

func (b *Benchmark) benchFile() string {
	r := b.GetTable().GetSelectedRowIndex()
	return ui.TrimCell(b.GetTable().SelectTable, r, 7)
}

// ----------------------------------------------------------------------------
// Helpers...

func fileToSubject(path string) string {
	tokens := strings.Split(path, "/")
	ee := strings.Split(tokens[len(tokens)-1], "_")
	return ee[0] + "/" + ee[1]
}

func benchDir(cfg *config.Config) string {
	ct, err := cfg.K9s.ActiveContext()
	if err != nil {
		log.Error().Err(err).Msgf("no active context located")
	}
	return filepath.Join(
		config.AppBenchmarksDir,
		data.SanitizeFileName(ct.ClusterName),
		data.SanitizeFileName(cfg.K9s.ActiveContextName()),
	)
}

func readBenchFile(cfg *config.Config, n string) (string, error) {
	data, err := os.ReadFile(filepath.Join(benchDir(cfg), n))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

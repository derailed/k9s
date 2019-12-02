package view

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	benchTitle  = "Benchmarks"
	resultTitle = "Benchmark Results"
)

// Bench represents a service benchmark results view.
type Bench struct {
	*Table

	details *Details
}

// NewBench returns a new viewer.
func NewBench(title, _ string, _ resource.List) ResourceViewer {
	return &Bench{
		Table:   NewTable(benchTitle),
		details: NewDetails(resultTitle),
	}
}

func (*Bench) SetContextFn(ContextFunc) {}

// Init initializes the viewer.
func (b *Bench) Init(ctx context.Context) error {
	log.Debug().Msgf(">>> Bench INIT")
	if err := b.Table.Init(ctx); err != nil {
		return err
	}
	b.SetBorderFocusColor(tcell.ColorSeaGreen)
	b.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
	b.SetColorerFn(render.Bench{}.ColorerFunc())
	b.bindKeys()

	b.details.SetTextColor(tcell.ColorSeaGreen)
	if err := b.details.Init(ctx); err != nil {
		return nil
	}

	b.Start()
	b.refresh()
	b.SetSortCol(b.NameColIndex()+7, 0, true)
	b.Refresh()
	b.Select(1, 0)

	return nil
}

// GVR returns a resource descriptor.
func (b *Bench) GVR() string {
	return "n/a"
}

// SetEnvFn sets k9s env vars.
func (b *Bench) SetEnvFn(EnvFunc) {}

// GetTable returns the table view.
func (b *Bench) GetTable() *Table { return b.Table }

// SetPath sets parent selector.
func (b *Bench) SetPath(s string) {}

// Start runs the refresh loop
func (b *Bench) Start() {
	log.Debug().Msgf(">>>> Bench START")
	var ctx context.Context

	ctx, b.cancelFn = context.WithCancel(context.Background())
	if err := b.watchBenchDir(ctx); err != nil {
		b.app.Flash().Errf("Unable to watch benchmarks directory %s", err)
	}
}

// List returns a resource list.
func (b *Bench) List() resource.List {
	return nil
}

func (b *Bench) refresh() {
	b.Update(b.hydrate())
	b.UpdateTitle()
}

func (b *Bench) bindKeys() {
	b.Actions().Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Enter", b.enterCmd, false),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", b.deleteCmd, false),
	})
}

func (b *Bench) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if b.SearchBuff().IsActive() {
		return b.filterCmd(evt)
	}

	if !b.RowSelected() {
		return nil
	}

	data, err := readBenchFile(b.app.Config, b.benchFile())
	if err != nil {
		b.app.Flash().Errf("Unable to load bench file %s", err)
		return nil
	}

	b.details.SetText(data)
	b.details.SetSubject(b.GetSelectedItem())
	b.app.inject(b.details)

	return nil
}

func (b *Bench) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.RowSelected() {
		return nil
	}

	sel, file := b.GetSelectedItem(), b.benchFile()
	dir := filepath.Join(perf.K9sBenchDir, b.app.Config.K9s.CurrentCluster)
	showModal(b.app.Content.Pages, fmt.Sprintf("Delete benchmark `%s?", file), func() {
		if err := os.Remove(filepath.Join(dir, file)); err != nil {
			b.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		b.app.Flash().Infof("Benchmark %s deleted!", sel)
	})

	return nil
}

func (b *Bench) benchFile() string {
	r := b.GetSelectedRowIndex()
	return ui.TrimCell(b.SelectTable, r, 7)
}

func (b *Bench) hydrate() render.TableData {
	ff, err := loadBenchDir(b.app.Config)
	if err != nil {
		b.app.Flash().Errf("Unable to read bench directory %s", err)
	}

	var re render.Bench
	data := render.TableData{
		Header:    re.Header(render.AllNamespaces),
		RowEvents: make(render.RowEvents, 0, 10),
		Namespace: render.AllNamespaces,
	}

	for _, f := range ff {
		bench := render.BenchInfo{
			File: f,
			Path: filepath.Join(benchDir(b.app.Config), f.Name()),
		}

		var row render.Row
		if err := re.Render(bench, render.AllNamespaces, &row); err != nil {
			log.Error().Err(err).Msg("Bench render failed")
			continue
		}
		data.RowEvents = append(data.RowEvents, render.RowEvent{
			Kind: render.EventAdd,
			Row:  row,
		})
	}

	return data
}

func (b *Bench) watchBenchDir(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("Bench event %#v", evt)
				b.app.QueueUpdateDraw(func() {
					b.refresh()
				})
			case err := <-w.Errors:
				log.Info().Err(err).Msg("Dir Watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msg("!!!! FS WATCHER DONE!!")
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing bench watched")
				}
				return
			}
		}
	}()

	return w.Add(benchDir(b.app.Config))
}

// ----------------------------------------------------------------------------
// Helpers...

func benchDir(cfg *config.Config) string {
	return filepath.Join(perf.K9sBenchDir, cfg.K9s.CurrentCluster)
}

func loadBenchDir(cfg *config.Config) ([]os.FileInfo, error) {
	return ioutil.ReadDir(benchDir(cfg))
}

func readBenchFile(cfg *config.Config, n string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(benchDir(cfg), n))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

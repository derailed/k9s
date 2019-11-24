package view

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/perf"
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

var (
	totalRx     = regexp.MustCompile(`Total:\s+([0-9.]+)\ssecs`)
	reqRx       = regexp.MustCompile(`Requests/sec:\s+([0-9.]+)`)
	okRx        = regexp.MustCompile(`\[2\d{2}\]\s+(\d+)\s+responses`)
	errRx       = regexp.MustCompile(`\[[4-5]\d{2}\]\s+(\d+)\s+responses`)
	toastRx     = regexp.MustCompile(`Error distribution`)
	benchHeader = resource.Row{"NAMESPACE", "NAME", "STATUS", "TIME", "REQ/S", "2XX", "4XX/5XX", "REPORT", "AGE"}
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

// Init initializes the viewer.
func (b *Bench) Init(ctx context.Context) error {
	log.Debug().Msgf(">>> Bench INIT")
	if err := b.Table.Init(ctx); err != nil {
		return err
	}
	b.SetBorderFocusColor(tcell.ColorSeaGreen)
	b.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
	b.SetColorerFn(benchColorer)
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

func (b *Bench) SetEnvFn(EnvFunc) {}
func (b *Bench) GetTable() *Table { return b.Table }

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

func (b *Bench) hydrate() resource.TableData {
	ff, err := loadBenchDir(b.app.Config)
	if err != nil {
		b.app.Flash().Errf("Unable to read bench directory %s", err)
	}

	data := initTable()
	for _, f := range ff {
		bench, err := readBenchFile(b.app.Config, f.Name())
		if err != nil {
			log.Error().Err(err).Msgf("Unable to load bench file %s", f.Name())
			continue
		}
		fields := make(resource.Row, len(benchHeader))
		if err := initRow(fields, f); err != nil {
			log.Error().Err(err).Msg("Load bench file")
			continue
		}
		augmentRow(fields, bench)
		data.Rows[f.Name()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func initRow(row resource.Row, f os.FileInfo) error {
	tokens := strings.Split(f.Name(), "_")
	if len(tokens) < 2 {
		return fmt.Errorf("Invalid file name %s", f.Name())
	}
	row[0] = tokens[0]
	row[1] = tokens[1]
	row[7] = f.Name()
	row[8] = time.Since(f.ModTime()).String()

	return nil
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

func initTable() resource.TableData {
	return resource.TableData{
		Header: benchHeader,
		Rows:   make(resource.RowEvents, 10),
		NumCols: map[string]bool{
			benchHeader[3]: true,
			benchHeader[4]: true,
			benchHeader[5]: true,
			benchHeader[6]: true,
		},
		Namespace: resource.AllNamespaces,
	}
}

func augmentRow(fields resource.Row, data string) {
	if len(data) == 0 {
		return
	}

	col := 2
	fields[col] = "pass"
	mf := toastRx.FindAllStringSubmatch(data, 1)
	if len(mf) > 0 {
		fields[col] = "fail"
	}
	col++

	mt := totalRx.FindAllStringSubmatch(data, 1)
	if len(mt) > 0 {
		fields[col] = mt[0][1]
	}
	col++

	mr := reqRx.FindAllStringSubmatch(data, 1)
	if len(mr) > 0 {
		fields[col] = mr[0][1]
	}
	col++

	ms := okRx.FindAllStringSubmatch(data, -1)
	fields[col] = countReq(ms)
	col++

	me := errRx.FindAllStringSubmatch(data, -1)
	fields[col] = countReq(me)
}

func countReq(rr [][]string) string {
	if len(rr) == 0 {
		return "0"
	}

	var sum int
	for _, m := range rr {
		if m, err := strconv.Atoi(string(m[1])); err == nil {
			sum += m
		}
	}
	return asNum(sum)
}

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

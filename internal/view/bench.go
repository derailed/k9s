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
	benchTitle = "Benchmarks"
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
	*MasterDetail

	cancelFn context.CancelFunc
}

// NewBench returns a new viewer.
func NewBench(_, _ string, _ resource.List) ResourceViewer {
	return &Bench{
		MasterDetail: NewMasterDetail(benchTitle, ""),
	}
}

// Init initializes the viewer.
func (b *Bench) Init(ctx context.Context) {
	b.MasterDetail.Init(ctx)
	b.keyBindings()

	tv := b.masterPage()
	tv.SetBorderFocusColor(tcell.ColorSeaGreen)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
	tv.SetColorerFn(benchColorer)

	dv := b.detailsPage()
	dv.setCategory("Bench")
	dv.SetTextColor(tcell.ColorSeaGreen)

	b.Start()
	b.refresh()
	tv.SetSortCol(tv.NameColIndex()+7, 0, true)
	tv.Refresh()
	tv.Select(1, 0)
}

func (b *Bench) Start() {
	var ctx context.Context

	ctx, b.cancelFn = context.WithCancel(context.Background())
	if err := b.watchBenchDir(ctx); err != nil {
		b.app.Flash().Errf("Unable to watch benchmarks directory %s", err)
	}
}

func (b *Bench) Stop() {
	if b.cancelFn != nil {
		b.cancelFn()
	}
}

func (b *Bench) Name() string {
	return "benchmarks"
}

func (b *Bench) setEnterFn(enterFn)            {}
func (b *Bench) setColorerFn(ui.ColorerFunc)   {}
func (b *Bench) setDecorateFn(decorateFn)      {}
func (b *Bench) setExtraActionsFn(ActionsFunc) {}

func (b *Bench) refresh() {
	tv := b.masterPage()
	tv.Update(b.hydrate())
	tv.UpdateTitle()
}

func (b *Bench) keyBindings() {
	aa := ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", b.app.PrevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Enter", b.enterCmd, false),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", b.deleteCmd, false),
	}
	b.masterPage().AddActions(aa)
}

func (b *Bench) getTitle() string {
	return benchTitle
}

func (b *Bench) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if b.masterPage().SearchBuff().IsActive() {
		return b.masterPage().filterCmd(evt)
	}

	if !b.masterPage().RowSelected() {
		return nil
	}

	data, err := readBenchFile(b.app.Config, b.benchFile())
	if err != nil {
		b.app.Flash().Errf("Unable to load bench file %s", err)
		return nil
	}
	vu := b.detailsPage()
	vu.SetText(data)
	vu.setTitle(b.masterPage().GetSelectedItem())
	b.showDetails()

	return nil
}

func (b *Bench) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.masterPage().RowSelected() {
		return nil
	}

	sel, file := b.masterPage().GetSelectedItem(), b.benchFile()
	dir := filepath.Join(perf.K9sBenchDir, b.app.Config.K9s.CurrentCluster)
	showModal(b.Pages, fmt.Sprintf("Delete benchmark `%s?", file), "master", func() {
		if err := os.Remove(filepath.Join(dir, file)); err != nil {
			b.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		b.app.Flash().Infof("Benchmark %s deleted!", sel)
	})

	return nil
}

func (b *Bench) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	b.showMaster()
	return nil
}

func (b *Bench) benchFile() string {
	r := b.masterPage().GetSelectedRowIndex()
	return ui.TrimCell(b.masterPage().Table, r, 7)
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
				w.Close()
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
	fields[col] = "0"
	if len(ms) > 0 {
		var sum int
		for _, m := range ms {
			if m, err := strconv.Atoi(string(m[1])); err == nil {
				sum += m
			}
		}
		fields[col] = asNum(sum)
	}
	col++

	me := errRx.FindAllStringSubmatch(data, -1)
	fields[col] = "0"
	if len(me) > 0 {
		var sum int
		for _, m := range me {
			if m, err := strconv.Atoi(string(m[1])); err == nil {
				sum += m
			}
		}
		fields[col] = asNum(sum)
	}
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

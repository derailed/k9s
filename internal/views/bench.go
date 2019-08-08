package views

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
	benchTitle    = "Benchmarks"
	benchTitleFmt = " [seagreen::b]%s([fuchsia::b]%d[fuchsia::-])[seagreen::-] "
)

var (
	totalRx     = regexp.MustCompile(`Total:\s+([0-9.]+)\ssecs`)
	reqRx       = regexp.MustCompile(`Requests/sec:\s+([0-9.]+)`)
	okRx        = regexp.MustCompile(`\[2\d{2}\]\s+(\d+)\s+responses`)
	errRx       = regexp.MustCompile(`\[[4-5]\d{2}\]\s+(\d+)\s+responses`)
	toastRx     = regexp.MustCompile(`Error distribution`)
	benchHeader = resource.Row{"NAMESPACE", "NAME", "STATUS", "TIME", "REQ/S", "2XX", "4XX/5XX", "REPORT", "AGE"}
)

type benchView struct {
	*masterDetail

	app *appView
}

func newBenchView(title string, app *appView, _ resource.List) resourceViewer {
	v := benchView{app: app}
	v.masterDetail = newMasterDetail(benchTitle, "", app, v.backCmd)
	v.keyBindings()

	return &v
}

// Init the view.
func (v *benchView) Init(ctx context.Context, ns string) {
	v.masterDetail.init(ctx, ns)

	tv := v.masterPage()
	tv.SetBorderFocusColor(tcell.ColorSeaGreen)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
	tv.SetColorerFn(benchColorer)

	dv := v.detailsPage()
	dv.setCategory("Bench")
	dv.SetTextColor(tcell.ColorSeaGreen)

	if err := v.watchBenchDir(ctx); err != nil {
		v.app.Flash().Errf("Unable to watch benchmarks directory %s", err)
	}

	v.refresh()
	tv.SetSortCol(tv.NameColIndex()+7, 0, true)
	tv.Refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
	v.app.SetHints(tv.Hints())
}

func (v *benchView) setEnterFn(enterFn)               {}
func (v *benchView) setColorerFn(ui.ColorerFunc)      {}
func (v *benchView) setDecorateFn(decorateFn)         {}
func (v *benchView) setExtraActionsFn(ui.ActionsFunc) {}

func (v *benchView) refresh() {
	tv := v.masterPage()
	tv.Update(v.hydrate())
	tv.UpdateTitle()
}

func (v *benchView) keyBindings() {
	aa := ui.KeyActions{
		ui.KeyP:        ui.NewKeyAction("Previous", v.app.prevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Enter", v.enterCmd, false),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", v.deleteCmd, false),
	}
	v.masterPage().SetActions(aa)
}

func (v *benchView) getTitle() string {
	return benchTitle
}

func (v *benchView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.masterPage()
		tv.SetSortCol(tv.NameColIndex()+col, 0, asc)
		tv.Refresh()

		return nil
	}
}

func (v *benchView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.masterPage().Cmd().IsActive() {
		return v.masterPage().filterCmd(evt)
	}

	if !v.masterPage().RowSelected() {
		return nil
	}

	data, err := readBenchFile(v.app.Config, v.benchFile())
	if err != nil {
		v.app.Flash().Errf("Unable to load bench file %s", err)
		return nil
	}
	vu := v.detailsPage()
	vu.SetText(data)
	vu.setTitle(v.masterPage().GetSelectedItem())
	v.showDetails()

	return nil
}

func (v *benchView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return nil
	}

	sel, file := v.masterPage().GetSelectedItem(), v.benchFile()
	dir := filepath.Join(perf.K9sBenchDir, v.app.Config.K9s.CurrentCluster)
	showModal(v.Pages, fmt.Sprintf("Delete benchmark `%s?", file), "master", func() {
		if err := os.Remove(filepath.Join(dir, file)); err != nil {
			v.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		v.app.Flash().Infof("Benchmark %s deleted!", sel)
	})

	return nil
}

func (v *benchView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.showMaster()
	return nil
}

func (v *benchView) benchFile() string {
	r := v.masterPage().GetSelectedRow()
	return ui.TrimCell(v.masterPage().Table, r, 7)
}

func (v *benchView) hints() ui.Hints {
	return v.CurrentPage().Item.(ui.Hinter).Hints()
}

func (v *benchView) hydrate() resource.TableData {
	ff, err := loadBenchDir(v.app.Config)
	if err != nil {
		v.app.Flash().Errf("Unable to read bench directory %s", err)
	}

	data := initTable()
	for _, f := range ff {
		bench, err := readBenchFile(v.app.Config, f.Name())
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

func (v *benchView) watchBenchDir(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("Bench event %#v", evt)
				v.app.QueueUpdateDraw(func() {
					v.refresh()
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

	return w.Add(benchDir(v.app.Config))
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

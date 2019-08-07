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
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
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
	*tview.Pages

	app          *appView
	cancel       context.CancelFunc
	selectedItem string
	selectedRow  int
	actions      ui.KeyActions
}

func newBenchView(_ string, app *appView, _ resource.List) resourceViewer {
	v := benchView{
		Pages:   tview.NewPages(),
		actions: make(ui.KeyActions),
		app:     app,
	}

	tv := newTableView(app, benchTitle)
	tv.SetSelectionChangedFunc(v.selChanged)
	tv.SetBorderFocusColor(tcell.ColorSeaGreen)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
	tv.SetColorerFn(benchColorer)
	tv.SetActiveNS("")
	v.AddPage("table", tv, true, true)

	details := newDetailsView(app, v.backCmd)
	details.setCategory("Bench")
	details.SetTextColor(app.Styles.FgColor())
	v.AddPage("details", details, true, false)

	v.registerActions()

	return &v
}

func (v *benchView) setEnterFn(enterFn)               {}
func (v *benchView) setColorerFn(ui.ColorerFunc)      {}
func (v *benchView) setDecorateFn(decorateFn)         {}
func (v *benchView) setExtraActionsFn(ui.ActionsFunc) {}

// Init the view.
func (v *benchView) Init(ctx context.Context, _ string) {
	if err := v.watchBenchDir(ctx); err != nil {
		v.app.Flash().Errf("Unable to watch benchmarks directory %s", err)
	}

	v.refresh()
	tv := v.getTV()
	tv.SetSortCol(tv.NameColIndex()+7, true)
	tv.Refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
}

func (v *benchView) refresh() {
	tv := v.getTV()
	tv.Update(v.hydrate())
	tv.UpdateTitle()
	v.selChanged(v.selectedRow, 0)
}

func (v *benchView) registerActions() {
	v.actions[ui.KeyP] = ui.NewKeyAction("Previous", v.app.prevCmd, false)
	v.actions[tcell.KeyEnter] = ui.NewKeyAction("Enter", v.enterCmd, false)
	v.actions[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", v.deleteCmd, false)

	vu := v.getTV()
	vu.SetActions(v.actions)
	v.app.SetHints(vu.Hints())
}

func (v *benchView) getTitle() string {
	return benchTitle
}

func (v *benchView) selChanged(r, c int) {
	tv := v.getTV()
	if r == 0 || tv.GetCell(r, 0) == nil {
		v.selectedItem = ""
		return
	}
	v.selectedRow = r
	v.selectedItem = ui.TrimCell(tv.Table, r, 7)
}

func (v *benchView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.SetSortCol(tv.NameColIndex()+col, asc)
		tv.Refresh()

		return nil
	}
}

func (v *benchView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.getTV().Cmd().IsActive() {
		return v.getTV().filterCmd(evt)
	}
	if v.selectedItem == "" {
		return nil
	}

	data, err := readBenchFile(v.app.Config, v.selectedItem)
	if err != nil {
		v.app.Flash().Errf("Unable to load bench file %s", err)
		return nil
	}
	vu := v.getDetails()
	vu.SetText(data)
	vu.setTitle(v.selectedItem)
	v.SwitchToPage("details")

	return nil
}

func (v *benchView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.selectedItem
	if sel == "" {
		return nil
	}

	dir := filepath.Join(K9sBenchDir, v.app.Config.K9s.CurrentCluster)
	showModal(v.Pages, fmt.Sprintf("Delete benchmark `%s?", sel), "table", func() {
		if err := os.Remove(filepath.Join(dir, sel)); err != nil {
			v.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		v.refresh()
		v.app.Flash().Infof("Benchmark %s deleted!", sel)
	})

	return nil
}

func (v *benchView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}
	v.SwitchToPage("table")

	return nil
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
		initRow(fields, f)
		augmentRow(fields, bench)
		data.Rows[f.Name()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func initRow(row resource.Row, f os.FileInfo) {
	tokens := strings.Split(f.Name(), "_")
	row[0] = tokens[0]
	row[1] = tokens[1]
	row[7] = f.Name()
	row[8] = time.Since(f.ModTime()).String()
}

func (v *benchView) getTV() *tableView {
	if vu, ok := v.GetPrimitive("table").(*tableView); ok {
		return vu
	}
	return nil
}

func (v *benchView) getDetails() *detailsView {
	if vu, ok := v.GetPrimitive("details").(*detailsView); ok {
		return vu
	}
	return nil
}

func (v *benchView) resetTitle1() {
	v.SetTitle(fmt.Sprintf(benchTitleFmt, benchTitle, v.getTV().GetRowCount()-1))
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
	return filepath.Join(K9sBenchDir, cfg.K9s.CurrentCluster)
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

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

	"github.com/derailed/k9s/internal/resource"
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
	current      igniter
	cancel       context.CancelFunc
	selectedItem string
	selectedRow  int
	actions      keyActions
}

func newBenchView(app *appView) *benchView {
	v := benchView{
		Pages:   tview.NewPages(),
		actions: make(keyActions),
		app:     app,
		current: app.content.GetPrimitive("main").(igniter),
	}

	tv := newTableView(app, benchTitle)
	{
		tv.SetSelectionChangedFunc(v.selChanged)
		tv.SetBorderFocusColor(tcell.ColorSeaGreen)
		tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorSeaGreen, tcell.AttrNone)
		tv.colorerFn = benchColorer
		tv.currentNS = ""
	}
	v.AddPage("table", tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)

	v.registerActions()

	return &v
}

// Init the view.
func (v *benchView) init(ctx context.Context, _ string) {
	if err := v.watchBenchDir(ctx); err != nil {
		log.Error().Err(err).Msg("Benchdir watch failed!")
		v.app.flash().errf("Unable to watch benchmarks directory %s", err)
	}

	tv := v.getTV()
	v.refresh()
	tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+7, true
	tv.refresh()
	v.app.SetFocus(tv)
}

func (v *benchView) refresh() {
	tv := v.getTV()
	tv.update(v.hydrate())
	tv.resetTitle()
	v.selChanged(v.selectedRow, 0)
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

func (v *benchView) registerActions() {
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)
	v.actions[tcell.KeyEnter] = newKeyAction("Enter", v.enterCmd, false)
	v.actions[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, false)

	vu := v.getTV()
	vu.setActions(v.actions)
	v.app.setHints(vu.hints())
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
	v.selectedItem = strings.TrimSpace(tv.GetCell(r, 7).Text)
	v.getTV().cmdBuff.setActive(false)
}

func (v *benchView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+col, asc
		tv.refresh()

		return nil
	}
}

func (v *benchView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.getTV().cmdBuff.isActive() {
		return v.getTV().filterCmd(evt)
	}

	sel := v.selectedItem
	if sel == "" {
		return nil
	}

	data, err := ioutil.ReadFile(filepath.Join(K9sBenchDir, sel))
	if err != nil {
		log.Error().Err(err).Msg("Read failed")
		v.app.flash().errf("Unable to load bench file %s", err)
		return nil
	}
	log.Debug().Msgf("Bench %v", string(data))
	vu := v.getDetails()
	vu.Clear()
	fmt.Fprintln(vu, string(data))
	{
		vu.setCategory("Bench")
		vu.setTitle(sel)
		vu.SetTextColor(tcell.ColorAqua)
		vu.ScrollToBeginning()
	}
	v.SwitchToPage("details")

	return nil
}

func (v *benchView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.selectedItem
	if sel == "" {
		return nil
	}

	showModal(v.Pages, fmt.Sprintf("Deleting `%s are you sure?", sel), "table", func() {
		if err := os.Remove(filepath.Join(K9sBenchDir, sel)); err != nil {
			v.app.flash().errf("Unable to delete file %s", err)
			log.Error().Err(err).Msg("Delete failed")
			return
		}
		v.refresh()
		v.app.flash().infof("Benchmark %s deleted!", sel)
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

func (v *benchView) hints() hints {
	return v.CurrentPage().Item.(hinter).hints()
}

func (v *benchView) hydrate() resource.TableData {
	cmds := helpCmds(v.app.conn())

	data := resource.TableData{
		Header:    benchHeader,
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: resource.AllNamespaces,
	}

	ff, err := ioutil.ReadDir(K9sBenchDir)
	if err != nil {
		log.Error().Err(err).Msg("Reading bench dir")
		v.app.flash().errf("Unable to read bench directory %s", err)
	}

	for _, f := range ff {
		bench, err := ioutil.ReadFile(filepath.Join(K9sBenchDir, f.Name()))
		if err != nil {
			continue
		}
		tokens := strings.Split(f.Name(), "_")
		fields := resource.Row{0: tokens[0], 1: tokens[1], 7: f.Name(), 8: time.Since(f.ModTime()).String()}
		augmentRow(fields, string(bench))
		data.Rows[f.Name()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
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
		fields[col] = strconv.Itoa(sum)
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
		fields[col] = strconv.Itoa(sum)
	}
}

func (v *benchView) resetTitle() {
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
				log.Info().Err(err).Msg("Skin watcher failed")
				return
			case <-ctx.Done():
				w.Close()
				return
			}
		}
	}()

	return w.Add(K9sBenchDir)
}

package views

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	dumpTitle    = "Screen Dumps"
	dumpTitleFmt = " [mediumvioletred::b]%s([fuchsia::b]%d[fuchsia::-])[mediumvioletred::-] "
)

var (
	dumpHeader = resource.Row{"NAME", "AGE"}
)

type dumpView struct {
	*tview.Pages

	app          *appView
	cancel       context.CancelFunc
	selectedItem string
	selectedRow  int
	actions      keyActions
}

func newDumpView(_ string, app *appView, _ resource.List) resourceViewer {
	v := dumpView{
		Pages:   tview.NewPages(),
		actions: make(keyActions),
		app:     app,
	}

	tv := newTableView(app, dumpTitle)
	{
		tv.SetSelectionChangedFunc(v.selChanged)
		tv.SetBorderFocusColor(tcell.ColorSteelBlue)
		tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
		tv.colorerFn = dumpColorer
		tv.currentNS = ""
	}
	v.AddPage("table", tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)
	v.registerActions()

	return &v
}

func (v *dumpView) setEnterFn(enterFn)          {}
func (v *dumpView) setColorerFn(colorerFn)      {}
func (v *dumpView) setDecorateFn(decorateFn)    {}
func (v *dumpView) setExtraActionsFn(actionsFn) {}

// Init the view.
func (v *dumpView) init(ctx context.Context, _ string) {
	if err := v.watchDumpDir(ctx); err != nil {
		v.app.flash().errf("Unable to watch dumpmarks directory %s", err)
	}

	tv := v.getTV()
	v.refresh()
	tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+1, true
	tv.refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
}

func (v *dumpView) refresh() {
	tv := v.getTV()
	tv.update(v.hydrate())
	tv.resetTitle()
	v.selChanged(v.selectedRow, 0)
}

func (v *dumpView) registerActions() {
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)
	v.actions[tcell.KeyEnter] = newKeyAction("Enter", v.enterCmd, true)
	v.actions[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, true)
	v.actions[tcell.KeyCtrlS] = newKeyAction("Save", noopCmd, false)

	vu := v.getTV()
	vu.setActions(v.actions)
	v.app.setHints(vu.hints())
}

func (v *dumpView) getTitle() string {
	return dumpTitle
}

func (v *dumpView) selChanged(r, c int) {
	log.Debug().Msgf("Selection changed %d:%c", r, c)
	tv := v.getTV()
	if r == 0 || tv.GetCell(r, 0) == nil {
		v.selectedItem = ""
		return
	}
	v.selectedRow = r
	v.selectedItem = strings.TrimSpace(tv.GetCell(r, 0).Text)
}

func (v *dumpView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+col, asc
		tv.refresh()

		return nil
	}
}

func (v *dumpView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.getTV().cmdBuff.isActive() {
		return v.getTV().filterCmd(evt)
	}
	sel := v.selectedItem
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	if !edit(true, v.app, filepath.Join(dir, sel)) {
		v.app.flash().err(errors.New("Failed to launch editor"))
	}

	return nil
}

func (v *dumpView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.selectedItem
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	showModal(v.Pages, fmt.Sprintf("Deleting `%s are you sure?", sel), "table", func() {
		if err := os.Remove(filepath.Join(dir, sel)); err != nil {
			v.app.flash().errf("Unable to delete file %s", err)
			return
		}
		v.refresh()
		v.app.flash().infof("ScreenDump file %s deleted!", sel)
	})

	return nil
}

func (v *dumpView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}
	v.SwitchToPage("table")
	return nil
}

func (v *dumpView) hints() hints {
	return v.CurrentPage().Item.(hinter).hints()
}

func (v *dumpView) hydrate() resource.TableData {
	data := resource.TableData{
		Header:    dumpHeader,
		Rows:      make(resource.RowEvents, 10),
		Namespace: resource.NotNamespaced,
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		v.app.flash().errf("Unable to read dump directory %s", err)
	}

	for _, f := range ff {
		fields := resource.Row{f.Name(), time.Since(f.ModTime()).String()}
		data.Rows[f.Name()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func (v *dumpView) resetTitle() {
	v.SetTitle(fmt.Sprintf(dumpTitleFmt, dumpTitle, v.getTV().GetRowCount()-1))
}

func (v *dumpView) watchDumpDir(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("Dump event %#v", evt)
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

	return w.Add(filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster))
}

func (v *dumpView) getTV() *tableView {
	if vu, ok := v.GetPrimitive("table").(*tableView); ok {
		return vu
	}
	return nil
}

func (v *dumpView) getDetails() *detailsView {
	if vu, ok := v.GetPrimitive("details").(*detailsView); ok {
		return vu
	}
	return nil
}

func noopCmd(*tcell.EventKey) *tcell.EventKey {
	return nil
}

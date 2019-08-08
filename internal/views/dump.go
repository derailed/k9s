package views

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	dumpTitle    = "Screen Dumps"
	dumpTitleFmt = " [mediumvioletred::b]%s([fuchsia::b]%d[fuchsia::-])[mediumvioletred::-] "
)

var (
	dumpHeader = resource.Row{"NAME", "AGE"}
)

type dumpView struct {
	*tview.Pages

	app    *appView
	cancel context.CancelFunc
}

func newDumpView(_ string, app *appView, _ resource.List) resourceViewer {
	v := dumpView{
		Pages: tview.NewPages(),
		app:   app,
	}

	tv := newTableView(app, dumpTitle)
	tv.SetBorderFocusColor(tcell.ColorSteelBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
	tv.SetColorerFn(dumpColorer)
	tv.SetActiveNS("")
	v.AddPage("table", tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)
	v.registerActions()

	return &v
}

func (v *dumpView) setEnterFn(enterFn)               {}
func (v *dumpView) setColorerFn(ui.ColorerFunc)      {}
func (v *dumpView) setDecorateFn(decorateFn)         {}
func (v *dumpView) setExtraActionsFn(ui.ActionsFunc) {}

// Init the view.
func (v *dumpView) Init(ctx context.Context, _ string) {
	if err := v.watchDumpDir(ctx); err != nil {
		v.app.Flash().Errf("Unable to watch dumpmarks directory %s", err)
	}

	tv := v.getTV()
	v.refresh()
	tv.SetSortCol(tv.NameColIndex()+1, 0, true)
	tv.Refresh()
	tv.SelectRow(1, true)
	v.app.SetFocus(tv)
}

func (v *dumpView) refresh() {
	tv := v.getTV()
	tv.Update(v.hydrate())
	tv.UpdateTitle()
}

func (v *dumpView) registerActions() {
	aa := ui.KeyActions{
		ui.KeyP:        ui.NewKeyAction("Previous", v.app.prevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Enter", v.enterCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", v.deleteCmd, true),
		tcell.KeyCtrlS: ui.NewKeyAction("Save", noopCmd, false),
	}

	tv := v.getTV()
	tv.SetActions(aa)
	v.app.SetHints(tv.Hints())
}

func (v *dumpView) getTitle() string {
	return dumpTitle
}

func (v *dumpView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.SetSortCol(tv.NameColIndex()+col, 0, asc)
		tv.Refresh()
		return nil
	}
}

func (v *dumpView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msg("Dump enter!")
	tv := v.getTV()
	if tv.Cmd().IsActive() {
		return tv.filterCmd(evt)
	}
	sel := tv.GetSelectedItem()
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster)
	if !edit(true, v.app, filepath.Join(dir, sel)) {
		v.app.Flash().Err(errors.New("Failed to launch editor"))
	}

	return nil
}

func (v *dumpView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.getTV().GetSelectedItem()
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster)
	showModal(v.Pages, fmt.Sprintf("Delete screen dump `%s?", sel), "table", func() {
		if err := os.Remove(filepath.Join(dir, sel)); err != nil {
			v.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		v.refresh()
		v.app.Flash().Infof("ScreenDump file %s deleted!", sel)
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

func (v *dumpView) hints() ui.Hints {
	return v.CurrentPage().Item.(ui.Hinter).Hints()
}

func (v *dumpView) hydrate() resource.TableData {
	data := resource.TableData{
		Header:    dumpHeader,
		Rows:      make(resource.RowEvents, 10),
		Namespace: resource.NotNamespaced,
	}

	dir := filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster)
	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		v.app.Flash().Errf("Unable to read dump directory %s", err)
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

	return w.Add(filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster))
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

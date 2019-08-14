package views

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	forwardTitle    = "Port Forwards"
	forwardTitleFmt = " [aqua::b]%s([fuchsia::b]%d[fuchsia::-])[aqua::-] "
	promptPage      = "prompt"
)

type forwardView struct {
	*tview.Pages

	app    *appView
	cancel context.CancelFunc
	bench  *perf.Benchmark
}

var _ resourceViewer = &forwardView{}

func newForwardView(ns string, app *appView, list resource.List) resourceViewer {
	v := forwardView{
		Pages: tview.NewPages(),
		app:   app,
	}

	tv := newTableView(app, forwardTitle)
	tv.SetBorderFocusColor(tcell.ColorDodgerBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	tv.SetColorerFn(forwardColorer)
	tv.SetActiveNS("")
	v.AddPage("table", tv, true, true)
	v.registerActions()

	return &v
}

func (v *forwardView) setEnterFn(enterFn)               {}
func (v *forwardView) setColorerFn(ui.ColorerFunc)      {}
func (v *forwardView) setDecorateFn(decorateFn)         {}
func (v *forwardView) setExtraActionsFn(ui.ActionsFunc) {}

// Init the view.
func (v *forwardView) Init(ctx context.Context, _ string) {
	path := ui.BenchConfig(v.app.Config.K9s.CurrentCluster)
	if err := watchFS(ctx, v.app, config.K9sHome, path, v.reload); err != nil {
		v.app.Flash().Errf("RuRoh! Unable to watch benchmarks directory %s : %s", config.K9sHome, err)
	}

	tv := v.getTV()
	v.refresh()
	tv.SetSortCol(tv.NameColIndex()+6, 0, true)
	tv.Refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
	v.app.SetHints(v.hints())
}

func (v *forwardView) getTV() *tableView {
	if vu, ok := v.GetPrimitive("table").(*tableView); ok {
		return vu
	}
	return nil
}

func (v *forwardView) reload() {
	path := ui.BenchConfig(v.app.Config.K9s.CurrentCluster)
	log.Debug().Msgf("Reloading Config %s", path)
	if err := v.app.Bench.Reload(path); err != nil {
		v.app.Flash().Err(err)
	}
	v.refresh()
}

func (v *forwardView) refresh() {
	tv := v.getTV()
	tv.Update(v.hydrate())
	v.app.SetFocus(tv)
	tv.UpdateTitle()
}

func (v *forwardView) registerActions() {
	tv := v.getTV()
	tv.SetActions(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Goto", v.gotoBenchCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench", v.benchCmd, true),
		tcell.KeyCtrlK: ui.NewKeyAction("Bench Stop", v.benchStopCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", v.deleteCmd, true),
		ui.KeySlash:    ui.NewKeyAction("Filter", tv.activateCmd, false),
		ui.KeyP:        ui.NewKeyAction("Previous", v.app.prevCmd, false),
		ui.KeyShiftP:   ui.NewKeyAction("Sort Ports", v.sortColCmd(2, true), true),
		ui.KeyShiftU:   ui.NewKeyAction("Sort URL", v.sortColCmd(4, true), true),
	})
}

func (v *forwardView) getTitle() string {
	return forwardTitle
}

func (v *forwardView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.SetSortCol(tv.NameColIndex()+col, 0, asc)
		v.refresh()

		return nil
	}
}

func (v *forwardView) gotoBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.gotoResource("be", true)

	return nil
}

func (v *forwardView) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.status(ui.FlashErr, "Benchmark Camceled!")
		v.bench.Cancel()
	}
	v.app.StatusReset()

	return nil
}

func (v *forwardView) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}

	if v.bench != nil {
		v.app.Flash().Err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	tv := v.getTV()
	r, _ := tv.GetSelection()
	cfg, co := defaultConfig(), ui.TrimCell(tv.Table, r, 2)
	if b, ok := v.app.Bench.Benchmarks.Containers[containerID(sel, co)]; ok {
		cfg = b
	}
	cfg.Name = sel

	base := ui.TrimCell(tv.Table, r, 4)
	var err error
	if v.bench, err = perf.NewBenchmark(base, cfg); err != nil {
		v.app.Flash().Errf("Bench failed %v", err)
		v.app.StatusReset()
		return nil
	}

	v.app.status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go v.runBenchmark()

	return nil
}

func (v *forwardView) runBenchmark() {
	v.bench.Run(v.app.Config.K9s.CurrentCluster, func() {
		log.Debug().Msg("Bench Completed!")
		v.app.QueueUpdate(func() {
			if v.bench.Canceled() {
				v.app.status(ui.FlashInfo, "Benchmark canceled")
			} else {
				v.app.status(ui.FlashInfo, "Benchmark Completed!")
				v.bench.Cancel()
			}
			v.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				v.app.QueueUpdate(func() { v.app.StatusReset() })
			}()
		})
	})
}

func (v *forwardView) getSelectedItem() string {
	tv := v.getTV()
	r, _ := tv.GetSelection()
	if r == 0 {
		return ""
	}
	return fwFQN(
		fqn(ui.TrimCell(tv.Table, r, 0), ui.TrimCell(tv.Table, r, 1)),
		ui.TrimCell(tv.Table, r, 2),
	)
}

func (v *forwardView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	tv := v.getTV()
	if !tv.Cmd().Empty() {
		tv.Cmd().Reset()
		return nil
	}

	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(v.Pages, fmt.Sprintf("Delete PortForward `%s?", sel), "table", func() {
		fw, ok := v.app.forwarders[sel]
		if !ok {
			log.Debug().Msgf("Unable to find forwarder %s", sel)
			return
		}
		fw.Stop()
		delete(v.app.forwarders, sel)

		log.Debug().Msgf("PortForwards after delete: %#v", v.app.forwarders)
		v.getTV().Update(v.hydrate())
		v.app.Flash().Infof("PortForward %s deleted!", sel)
	})

	return nil
}

func (v *forwardView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	tv := v.getTV()
	if tv.Cmd().IsActive() {
		tv.Cmd().Reset()
	} else {
		v.app.inject(v.app.Frame().GetPrimitive("main").(ui.Igniter))
	}

	return nil
}

func (v *forwardView) hints() ui.Hints {
	return v.getTV().Hints()
}

func (v *forwardView) hydrate() resource.TableData {
	data := initHeader(len(v.app.forwarders))
	dc, dn := v.app.Bench.Benchmarks.Defaults.C, v.app.Bench.Benchmarks.Defaults.N
	for _, f := range v.app.forwarders {
		c, n, cfg := loadConfig(dc, dn, containerID(f.Path(), f.Container()), v.app.Bench.Benchmarks.Containers)

		ports := strings.Split(f.Ports()[0], ":")
		ns, na := namespaced(f.Path())
		fields := resource.Row{
			ns,
			na,
			f.Container(),
			strings.Join(f.Ports(), ","),
			urlFor(cfg, f.Container(), ports[0]),
			asNum(c),
			asNum(n),
			f.Age(),
		}
		data.Rows[f.Path()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func (v *forwardView) resetTitle() {
	v.SetTitle(fmt.Sprintf(forwardTitleFmt, forwardTitle, v.getTV().GetRowCount()-1))
}

// ----------------------------------------------------------------------------
// Helpers...

func defaultConfig() config.BenchConfig {
	return config.BenchConfig{
		C: config.DefaultC,
		N: config.DefaultN,
		HTTP: config.HTTP{
			Method: config.DefaultMethod,
			Path:   "/",
		},
	}
}

func initHeader(rows int) resource.TableData {
	return resource.TableData{
		Header:    resource.Row{"NAMESPACE", "NAME", "CONTAINER", "PORTS", "URL", "C", "N", "AGE"},
		NumCols:   map[string]bool{"C": true, "N": true},
		Rows:      make(resource.RowEvents, rows),
		Namespace: resource.AllNamespaces,
	}
}

func loadConfig(dc, dn int, id string, cc map[string]config.BenchConfig) (int, int, config.BenchConfig) {
	c, n := dc, dn
	cfg, ok := cc[id]
	if !ok {
		return c, n, cfg
	}

	if cfg.C != 0 {
		c = cfg.C
	}
	if cfg.N != 0 {
		n = cfg.N
	}

	return c, n, cfg
}

func showModal(pv *tview.Pages, msg, back string, ok func()) {
	m := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(func(_ int, b string) {
			if b == "OK" {
				ok()
			}
			dismissModal(pv, back)
		})
	m.SetTitle("<Delete Benchmark>")
	pv.AddPage(promptPage, m, false, false)
	pv.ShowPage(promptPage)
}

func dismissModal(pv *tview.Pages, page string) {
	pv.RemovePage(promptPage)
	pv.SwitchToPage(page)
}

func watchFS(ctx context.Context, app *appView, dir, file string, cb func()) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("FS %s event %v", file, evt.Name)
				if file == "" || evt.Name == file {
					log.Debug().Msgf("Capuring Event %#v", evt)
					app.QueueUpdateDraw(func() {
						cb()
					})
				}
			case err := <-w.Errors:
				log.Info().Err(err).Msgf("FS %s watcher failed", dir)
				return
			case <-ctx.Done():
				log.Debug().Msgf("<<FS %s WATCHER DONE>>", dir)
				w.Close()
				return
			}
		}
	}()

	return w.Add(dir)
}

package views

import (
	"context"
	"errors"
	"fmt"
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
	forwardTitle    = "Port Forwards"
	forwardTitleFmt = " [aqua::b]%s([fuchsia::b]%d[fuchsia::-])[aqua::-] "
	promptPage      = "prompt"
)

type forwardView struct {
	*tview.Pages

	app     *appView
	current igniter
	cancel  context.CancelFunc
	bench   *benchmark
}

func newForwardView(app *appView) *forwardView {
	v := forwardView{
		Pages: tview.NewPages(),
		app:   app,
	}

	tv := newTableView(app, forwardTitle)
	tv.SetBorderFocusColor(tcell.ColorDodgerBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	tv.colorerFn = forwardColorer
	tv.currentNS = ""
	v.AddPage("table", tv, true, true)

	v.current = app.content.GetPrimitive("main").(igniter)
	v.registerActions()

	return &v
}

// Init the view.
func (v *forwardView) init(ctx context.Context, _ string) {
	if err := watchFS(ctx, v.app, config.K9sHome, config.K9sBenchmarks, v.reload); err != nil {
		log.Error().Err(err).Msg("Benchdir watch failed!")
		v.app.flash().errf("Unable to watch benchmarks directory %s", err)
	}

	tv := v.getTV()
	v.refresh()
	tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+6, true
	tv.refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
}

func (v *forwardView) getTV() *tableView {
	if vu, ok := v.GetPrimitive("table").(*tableView); ok {
		return vu
	}
	return nil
}

func (v *forwardView) reload() {
	if err := v.app.bench.Reload(); err != nil {
		log.Error().Err(err).Msg("Bench config reload")
		v.app.flash().err(err)
	}
	v.refresh()
}

func (v *forwardView) refresh() {
	tv := v.getTV()
	tv.update(v.hydrate())
	v.app.SetFocus(tv)
	tv.resetTitle()
}

func (v *forwardView) registerActions() {
	tv := v.getTV()
	tv.actions[tcell.KeyCtrlB] = newKeyAction("Bench", v.benchCmd, true)
	tv.actions[KeyAltB] = newKeyAction("Bench Stop", v.benchStopCmd, true)
	tv.actions[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, true)
	tv.actions[KeySlash] = newKeyAction("Filter", tv.activateCmd, false)
	tv.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)
	tv.actions[KeyShiftP] = newKeyAction("Sort Ports", v.sortColCmd(2, true), true)
	tv.actions[KeyShiftU] = newKeyAction("Sort URL", v.sortColCmd(4, true), true)
}

func (v *forwardView) getTitle() string {
	return forwardTitle
}

func (v *forwardView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+col, asc
		v.refresh()

		return nil
	}
}

func (v *forwardView) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.flash().info("Benchmark canceled!")
		v.app.status(flashErr, "Benchmark Camceled!")
		v.bench.cancel()
	}
	v.app.statusReset()

	return nil
}

func (v *forwardView) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}
	if v.bench != nil {
		v.app.flash().err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	tv := v.getTV()
	r, _ := tv.GetSelection()
	c, n := v.app.bench.Benchmarks.Defaults.C, v.app.bench.Benchmarks.Defaults.N
	m, url := config.DefaultMethod, strings.TrimSpace(tv.GetCell(r, 4).Text)
	container := strings.TrimSpace(tv.GetCell(r, 2).Text)
	if b, ok := v.app.bench.Benchmarks.Containers[container]; ok {
		c, n = b.C, b.N
	}

	cfg := benchConfig{
		Path:   sel,
		Method: m,
		URL:    url,
		C:      c,
		N:      n,
	}
	log.Debug().Msgf(">>>>> BENCHCONFIG %#v", cfg)
	var err error
	if v.bench, err = newBenchmark(cfg); err != nil {
		log.Error().Err(err).Msg("Bench failed!")
		v.app.flash().errf("Bench failed %v", err)
		v.app.statusReset()
		v.bench = nil
		return nil
	}

	v.app.status(flashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go v.bench.run(v.app.config.K9s.CurrentCluster, func() {
		log.Debug().Msg("Bench Completed!")
		v.app.QueueUpdate(func() {
			if v.bench.canceled {
				v.app.status(flashInfo, "Benchmark canceled")
			} else {
				v.app.flash().infof("Benchmark for %s is done!", sel)
				v.app.status(flashInfo, "Benchmark Completed!")
				v.bench.cancel()
			}
			v.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				v.app.QueueUpdate(func() {
					v.app.statusReset()
				})
			}()
		})
	})

	return nil
}

func (v *forwardView) getSelectedItem() string {
	tv := v.getTV()
	r, _ := tv.GetSelection()
	if r == 0 {
		return ""
	}

	return fqn(strings.TrimSpace(tv.GetCell(r, 0).Text), strings.TrimSpace(tv.GetCell(r, 1).Text))
}

func (v *forwardView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	tv := v.getTV()
	if !tv.cmdBuff.empty() {
		tv.cmdBuff.reset()
		return nil
	}
	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(v.Pages, fmt.Sprintf("Deleting `%s are you sure?", sel), "table", func() {
		index := -1
		for i, f := range v.app.forwarders {
			if sel == f.Path() {
				index = i
			}
		}
		if index == -1 {
			return
		}
		v.app.forwarders[index].Stop()
		if index == 0 && len(v.app.forwarders) == 1 {
			v.app.forwarders = []forwarder{}
		} else {
			v.app.forwarders = append(v.app.forwarders[:index], v.app.forwarders[index+1:]...)
		}
		log.Debug().Msgf("PortForwards after delete: %#v", v.app.forwarders)
		v.getTV().update(v.hydrate())
		v.app.flash().infof("PortForward %s deleted!", sel)
	})

	return nil
}

func (v *forwardView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	tv := v.getTV()
	if tv.cmdBuff.isActive() {
		tv.cmdBuff.reset()
	} else {
		v.app.inject(v.current)
	}

	return nil
}

func (v *forwardView) hints() hints {
	return v.getTV().actions.toHints()
}

func (v *forwardView) hydrate() resource.TableData {
	data := resource.TableData{
		Header:    resource.Row{"NAMESPACE", "NAME", "CONTAINER", "PORTS", "URL", "C", "N", "AGE"},
		NumCols:   map[string]bool{"C": true, "N": true},
		Rows:      make(resource.RowEvents, len(v.app.forwarders)),
		Namespace: resource.AllNamespaces,
	}

	dc, dn := v.app.bench.Benchmarks.Defaults.C, v.app.bench.Benchmarks.Defaults.N
	for _, f := range v.app.forwarders {
		c, n := dc, dn
		if b, ok := v.app.bench.Benchmarks.Containers[f.Container()]; ok {
			c, n = b.C, b.N
		}

		ports := strings.Split(f.Ports()[0], ":")
		ns, na := namespaced(f.Path())
		fields := resource.Row{
			ns,
			na,
			f.Container(),
			strings.Join(f.Ports(), ","),
			urlFor(v.app.bench.Benchmarks, f.Container(), ports[0]),
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

func urlFor(cfg *config.Benchmarks, co, port string) string {
	path := "/"
	if b, ok := cfg.Containers[co]; ok {
		if b.Path != "" {
			path = b.Path
		}
	}

	return "http://localhost" + ":" + port + path
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
	m.SetTitle("<Confirm>")
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

	path := filepath.Join(dir, file)
	if file == "" {
		path = ""
	}
	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("Event %#v", evt)
				if file == "" || evt.Name == path {
					log.Debug().Msgf("FS %s event %v", dir, evt)
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

package view

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
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

// PortForward presents active portforward viewer.
type PortForward struct {
	*ui.Pages

	cancelFn context.CancelFunc
	bench    *perf.Benchmark
	app      *App
}

// NewPortForward returns a new viewer.
func NewPortForward(title, _ string, list resource.List) ResourceViewer {
	return &PortForward{
		Pages: ui.NewPages(),
	}
}

// Init the view.
func (p *PortForward) Init(ctx context.Context) {
	p.app = ctx.Value(ui.KeyApp).(*App)

	tv := NewTable(forwardTitle)
	tv.Init(ctx)
	tv.SetBorderFocusColor(tcell.ColorDodgerBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	tv.SetColorerFn(forwardColorer)
	tv.SetActiveNS("")
	tv.SetSortCol(tv.NameColIndex()+6, 0, true)
	tv.Select(1, 0)
	p.Push(tv)

	p.registerActions()
	p.Start()
	p.refresh()
}

func (p *PortForward) Start() {
	path := ui.BenchConfig(p.app.Config.K9s.CurrentCluster)
	var ctx context.Context
	ctx, p.cancelFn = context.WithCancel(context.Background())
	if err := watchFS(ctx, p.app, config.K9sHome, path, p.reload); err != nil {
		p.app.Flash().Errf("RuRoh! Unable to watch benchmarks directory %s : %s", config.K9sHome, err)
	}
}

func (p *PortForward) Stop() {}

func (p *PortForward) Name() string {
	return "portForwards"
}

func (p *PortForward) masterPage() *Table {
	return p.GetPrimitive("table").(*Table)
}

func (p *PortForward) setEnterFn(enterFn)            {}
func (p *PortForward) setColorerFn(ui.ColorerFunc)   {}
func (p *PortForward) setDecorateFn(decorateFn)      {}
func (p *PortForward) setExtraActionsFn(ActionsFunc) {}

func (p *PortForward) getTV() *Table {
	if vu, ok := p.GetPrimitive("table").(*Table); ok {
		return vu
	}
	return nil
}

func (p *PortForward) reload() {
	path := ui.BenchConfig(p.app.Config.K9s.CurrentCluster)
	log.Debug().Msgf("Reloading Config %s", path)
	if err := p.app.Bench.Reload(path); err != nil {
		p.app.Flash().Err(err)
	}
	p.refresh()
}

func (p *PortForward) refresh() {
	tv := p.getTV()
	tv.Update(p.hydrate())
	p.app.SetFocus(tv)
	tv.UpdateTitle()
}

func (p *PortForward) registerActions() {
	tv := p.getTV()
	tv.AddActions(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Goto", p.gotoBenchCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench", p.benchCmd, true),
		tcell.KeyCtrlK: ui.NewKeyAction("Bench Stop", p.benchStopCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", p.deleteCmd, true),
		ui.KeySlash:    ui.NewKeyAction("Filter", tv.activateCmd, false),
		ui.KeyP:        ui.NewKeyAction("Previous", p.app.PrevCmd, false),
		ui.KeyShiftP:   ui.NewKeyAction("Sort Ports", p.sortColCmd(2, true), false),
		ui.KeyShiftU:   ui.NewKeyAction("Sort URL", p.sortColCmd(4, true), false),
	})
}

func (p *PortForward) getTitle() string {
	return forwardTitle
}

func (p *PortForward) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := p.getTV()
		tv.SetSortCol(tv.NameColIndex()+col, 0, asc)
		p.refresh()

		return nil
	}
}

func (p *PortForward) gotoBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	p.app.gotoResource("be", true)

	return nil
}

func (p *PortForward) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.bench != nil {
		log.Debug().Msg(">>> Benchmark cancelFned!!")
		p.app.status(ui.FlashErr, "Benchmark Camceled!")
		p.bench.Cancel()
	}
	p.app.StatusReset()

	return nil
}

func (p *PortForward) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := p.getSelectedItem()
	if sel == "" {
		return nil
	}

	if p.bench != nil {
		p.app.Flash().Err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	tv := p.getTV()
	r, _ := tv.GetSelection()
	cfg, co := defaultConfig(), ui.TrimCell(tv.Table, r, 2)
	if b, ok := p.app.Bench.Benchmarks.Containers[containerID(sel, co)]; ok {
		cfg = b
	}
	cfg.Name = sel

	base := ui.TrimCell(tv.Table, r, 4)
	var err error
	if p.bench, err = perf.NewBenchmark(base, cfg); err != nil {
		p.app.Flash().Errf("Bench failed %v", err)
		p.app.StatusReset()
		return nil
	}

	p.app.status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go p.runBenchmark()

	return nil
}

func (p *PortForward) runBenchmark() {
	p.bench.Run(p.app.Config.K9s.CurrentCluster, func() {
		log.Debug().Msg("Bench Completed!")
		p.app.QueueUpdate(func() {
			if p.bench.Canceled() {
				p.app.status(ui.FlashInfo, "Benchmark cancelFned")
			} else {
				p.app.status(ui.FlashInfo, "Benchmark Completed!")
				p.bench.Cancel()
			}
			p.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				p.app.QueueUpdate(func() { p.app.StatusReset() })
			}()
		})
	})
}

func (p *PortForward) getSelectedItem() string {
	tv := p.getTV()
	r, _ := tv.GetSelection()
	if r == 0 {
		return ""
	}
	return fwFQN(
		fqn(ui.TrimCell(tv.Table, r, 0), ui.TrimCell(tv.Table, r, 1)),
		ui.TrimCell(tv.Table, r, 2),
	)
}

func (p *PortForward) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	tv := p.getTV()
	if !tv.SearchBuff().Empty() {
		tv.SearchBuff().Reset()
		return nil
	}

	sel := p.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(p.Pages, fmt.Sprintf("Delete PortForward `%s?", sel), "table", func() {
		fw, ok := p.app.forwarders[sel]
		if !ok {
			log.Debug().Msgf("Unable to find forwarder %s", sel)
			return
		}
		fw.Stop()
		delete(p.app.forwarders, sel)

		log.Debug().Msgf("PortForwards after delete: %#v", p.app.forwarders)
		p.getTV().Update(p.hydrate())
		p.app.Flash().Infof("PortForward %s deleted!", sel)
	})

	return nil
}

func (p *PortForward) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.cancelFn != nil {
		p.cancelFn()
	}

	tv := p.getTV()
	if tv.SearchBuff().IsActive() {
		tv.SearchBuff().Reset()
	} else {
		p.app.inject(p.app.Content.GetPrimitive("main").(model.Component))
	}

	return nil
}

func (p *PortForward) Hints() model.MenuHints {
	return p.getTV().Hints()
}

func (p *PortForward) hydrate() resource.TableData {
	data := initHeader(len(p.app.forwarders))
	dc, dn := p.app.Bench.Benchmarks.Defaults.C, p.app.Bench.Benchmarks.Defaults.N
	for _, f := range p.app.forwarders {
		c, n, cfg := loadConfig(dc, dn, containerID(f.Path(), f.Container()), p.app.Bench.Benchmarks.Containers)

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

func (p *PortForward) resetTitle() {
	p.SetTitle(fmt.Sprintf(forwardTitleFmt, forwardTitle, p.getTV().GetRowCount()-1))
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

func showModal(p *ui.Pages, msg, back string, ok func()) {
	m := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(func(_ int, b string) {
			if b == "OK" {
				ok()
			}
			dismissModal(p, back)
		})
	m.SetTitle("<Delete Benchmark>")
	p.AddPage(promptPage, m, false, false)
	p.ShowPage(promptPage)
}

func dismissModal(p *ui.Pages, page string) {
	p.RemovePage(promptPage)
	p.SwitchToPage(page)
}

func watchFS(ctx context.Context, app *App, dir, file string, cb func()) error {
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

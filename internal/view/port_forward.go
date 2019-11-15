package view

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
	portForwardTitle = "PortForwards"
	promptPage       = "prompt"
)

// PortForward presents active portforward viewer.
type PortForward struct {
	*MasterDetail

	cancelFn context.CancelFunc
	bench    *perf.Benchmark
	app      *App
}

// NewPortForward returns a new viewer.
func NewPortForward(title, gvr string, list resource.List) ResourceViewer {
	return &PortForward{
		MasterDetail: NewMasterDetail(portForwardTitle, ""),
	}
}

// Init the view.
func (p *PortForward) Init(ctx context.Context) {
	p.app = mustExtractApp(ctx)
	p.MasterDetail.Init(ctx)
	p.registerActions()

	tv := p.masterPage()
	tv.SetBorderFocusColor(tcell.ColorDodgerBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	tv.SetColorerFn(forwardColorer)
	tv.SetActiveNS("")
	tv.SetSortCol(tv.NameColIndex()+6, 0, true)
	tv.Select(1, 0)

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
	return portForwardTitle
}

func (p *PortForward) setEnterFn(enterFn)            {}
func (p *PortForward) setColorerFn(ui.ColorerFunc)   {}
func (p *PortForward) setDecorateFn(decorateFn)      {}
func (p *PortForward) setExtraActionsFn(ActionsFunc) {}

func (p *PortForward) reload() {
	path := ui.BenchConfig(p.app.Config.K9s.CurrentCluster)
	log.Debug().Msgf("Reloading Config %s", path)
	if err := p.app.Bench.Reload(path); err != nil {
		p.app.Flash().Err(err)
	}
	p.refresh()
}

func (p *PortForward) refresh() {
	tv := p.masterPage()
	tv.Update(p.hydrate())
	p.app.SetFocus(tv)
	tv.UpdateTitle()
}

func (p *PortForward) registerActions() {
	tv := p.masterPage()
	tv.AddActions(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Benchmarks", p.gotoBenchCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench", p.benchCmd, true),
		tcell.KeyCtrlK: ui.NewKeyAction("Bench Stop", p.benchStopCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", p.deleteCmd, true),
		ui.KeySlash:    ui.NewKeyAction("Filter", tv.activateCmd, false),
		tcell.KeyEsc:   ui.NewKeyAction("Back", p.app.PrevCmd, false),
		ui.KeyShiftP:   ui.NewKeyAction("Sort Ports", p.sortColCmd(2, true), false),
		ui.KeyShiftU:   ui.NewKeyAction("Sort URL", p.sortColCmd(4, true), false),
	})
}

func (p *PortForward) gotoBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	p.app.gotoResource("be")

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

	tv := p.masterPage()
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
	tv := p.masterPage()
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
	tv := p.masterPage()
	if !tv.SearchBuff().Empty() {
		tv.SearchBuff().Reset()
		return nil
	}

	sel := p.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(p.Pages, fmt.Sprintf("Delete PortForward `%s?", sel), func() {
		fw, ok := p.app.forwarders[sel]
		if !ok {
			log.Debug().Msgf("Unable to find forwarder %s", sel)
			return
		}
		fw.Stop()
		delete(p.app.forwarders, sel)

		log.Debug().Msgf("PortForwards after delete: %#v", p.app.forwarders)
		p.masterPage().Update(p.hydrate())
		p.app.Flash().Infof("PortForward %s deleted!", sel)
	})

	return nil
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
			urlFor(cfg, ports[0]),
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

func showModal(p *ui.Pages, msg string, ok func()) {
	m := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(func(_ int, b string) {
			if b == "OK" {
				ok()
			}
			dismissModal(p)
		})
	m.SetTitle("<Delete Benchmark>")
	p.AddPage(promptPage, m, false, false)
	p.ShowPage(promptPage)
}

func dismissModal(p *ui.Pages) {
	p.RemovePage(promptPage)
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
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing portforward watcher")
				}
				return
			}
		}
	}()

	return w.Add(dir)
}

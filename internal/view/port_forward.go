package view

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/render"
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
	ResourceViewer

	bench *perf.Benchmark
}

// NewPortForward returns a new viewer.
func NewPortForward(gvr dao.GVR) ResourceViewer {
	p := PortForward{
		ResourceViewer: NewBrowser(gvr),
	}
	p.GetTable().SetBorderFocusColor(tcell.ColorDodgerBlue)
	p.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	p.GetTable().SetColorerFn(render.PortForward{}.ColorerFunc())
	p.GetTable().SetSortCol(p.GetTable().NameColIndex()+6, 0, true)
	p.SetContextFn(p.portForwardContext)
	p.SetBindKeysFn(p.bindKeys)

	return &p
}

func (p *PortForward) portForwardContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyBenchCfg, p.App().Bench)
}

// BOZO!!
// // Start runs the refresh loop.
// func (p *PortForward) Start() {
// 	path := ui.BenchConfig(p.App().Config.K9s.CurrentCluster)
// 	var ctx context.Context
// 	ctx, p.cancelFn = context.WithCancel(context.Background())
// 	if err := watchFS(ctx, p.App(), config.K9sHome, path, p.reload); err != nil {
// 		p.App().Flash().Errf("RuRoh! Unable to watch benchmarks directory %s : %s", config.K9sHome, err)
// 	}
// }

// // Name returns the component name.
// func (p *PortForward) Name() string {
// 	return portForwardTitle
// }

// func (p *PortForward) reload() {
// 	path := ui.BenchConfig(p.App().Config.K9s.CurrentCluster)
// 	log.Debug().Msgf("Reloading Config %s", path)
// 	if err := p.App().Bench.Reload(path); err != nil {
// 		p.App().Flash().Err(err)
// 	}
// 	p.refresh()
// }

// func (p *PortForward) refresh() {
// 	p.Update(p.hydrate())
// 	p.App().SetFocus(p)
// 	p.UpdateTitle()
// }

func (p *PortForward) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Benchmarks", p.showBenchCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench", p.benchCmd, true),
		tcell.KeyCtrlK: ui.NewKeyAction("Bench Stop", p.benchStopCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", p.deleteCmd, true),
		// ui.KeySlash:    ui.NewKeyAction("Filter", p.activateCmd, false),
		tcell.KeyEsc: ui.NewKeyAction("Back", p.App().PrevCmd, false),
		ui.KeyShiftP: ui.NewKeyAction("Sort Ports", p.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftU: ui.NewKeyAction("Sort URL", p.GetTable().SortColCmd(4, true), false),
	})
}

func (p *PortForward) showBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	p.App().inject(NewBenchmark("benchmarks"))

	return nil
}

func (p *PortForward) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.bench != nil {
		log.Debug().Msg(">>> Benchmark cancelFned!!")
		p.App().status(ui.FlashErr, "Benchmark Camceled!")
		p.bench.Cancel()
	}
	p.App().StatusReset()

	return nil
}

func (p *PortForward) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := p.getSelectedItem()
	if sel == "" {
		return nil
	}

	if p.bench != nil {
		p.App().Flash().Err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	r, _ := p.GetTable().GetSelection()
	cfg, co := defaultConfig(), ui.TrimCell(p.GetTable().SelectTable, r, 2)
	if b, ok := p.App().Bench.Benchmarks.Containers[containerID(sel, co)]; ok {
		cfg = b
	}
	cfg.Name = sel

	base := ui.TrimCell(p.GetTable().SelectTable, r, 4)
	var err error
	if p.bench, err = perf.NewBenchmark(base, cfg); err != nil {
		p.App().Flash().Errf("Bench failed %v", err)
		p.App().StatusReset()
		return nil
	}

	p.App().status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go p.runBenchmark()

	return nil
}

func (p *PortForward) runBenchmark() {
	p.bench.Run(p.App().Config.K9s.CurrentCluster, func() {
		log.Debug().Msg("Bench Completed!")
		p.App().QueueUpdate(func() {
			if p.bench.Canceled() {
				p.App().status(ui.FlashInfo, "Benchmark cancelFned")
			} else {
				p.App().status(ui.FlashInfo, "Benchmark Completed!")
				p.bench.Cancel()
			}
			p.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				p.App().QueueUpdate(func() { p.App().StatusReset() })
			}()
		})
	})
}

func (p *PortForward) getSelectedItem() string {
	r, _ := p.GetTable().GetSelection()
	if r == 0 {
		return ""
	}
	return fwFQN(
		fqn(ui.TrimCell(p.GetTable().SelectTable, r, 0), ui.TrimCell(p.GetTable().SelectTable, r, 1)),
		ui.TrimCell(p.GetTable().SelectTable, r, 2),
	)
}

func (p *PortForward) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.GetTable().SearchBuff().Empty() {
		p.GetTable().SearchBuff().Reset()
		return nil
	}

	sel := p.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(p.App().Content.Pages, fmt.Sprintf("Delete PortForward `%s?", sel), func() {
		p.App().factory.DeleteForwarder(sel)
		p.App().Flash().Infof("PortForward %s(%d) deleted!", sel)
		p.GetTable().Refresh()
	})

	return nil
}

// func (p *PortForward) hydrate() render.TableData {
// 	var re render.Forward

// 	data := render.TableData{
// 		Header:    re.Header(render.AllNamespaces),
// 		RowEvents: make(render.RowEvents, 0, len(p.App().forwarders)),
// 		Namespace: render.AllNamespaces,
// 	}

// 	containers := p.App().Bench.Benchmarks.Containers
// 	for _, f := range p.App().forwarders {
// 		fqn := containerID(f.Path(), f.Container())
// 		cfg := benchCfg{
// 			c: p.App().Bench.Benchmarks.Defaults.C,
// 			n: p.App().Bench.Benchmarks.Defaults.N,
// 		}
// 		if config, ok := containers[fqn]; ok {
// 			cfg.c, cfg.n = config.C, config.N
// 			cfg.host, cfg.path = config.HTTP.Host, config.HTTP.Path
// 		}

// 		var row render.Row
// 		fwd := forwarder{
// 			Forwarder:         f,
// 			BenchConfigurator: cfg,
// 		}
// 		if err := re.Render(fwd, render.AllNamespaces, &row); err != nil {
// 			log.Error().Err(err).Msgf("PortForward render failed")
// 			continue
// 		}
// 		data.RowEvents = append(data.RowEvents, render.RowEvent{Kind: render.EventAdd, Row: row})
// 	}

// 	return data
// }

// ----------------------------------------------------------------------------
// Helpers...

// var _ render.PortForwarder = forwarder{}

// type forwarder struct {
// 	render.Forwarder
// 	render.BenchConfigurator
// }

// type benchCfg struct {
// 	c, n       int
// 	host, path string
// }

// func (b benchCfg) C() int           { return b.c }
// func (b benchCfg) N() int           { return b.n }
// func (b benchCfg) Host() string     { return b.host }
// func (b benchCfg) HttpPath() string { return b.path }

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
					log.Debug().Msgf("Capturing Event %#v", evt)
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

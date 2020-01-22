package view

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const promptPage = "prompt"

// PortForward presents active portforward viewer.
type PortForward struct {
	ResourceViewer

	bench *perf.Benchmark
}

// NewPortForward returns a new viewer.
func NewPortForward(gvr client.GVR) ResourceViewer {
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

func (p *PortForward) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("View Benchmarks", p.showBenchCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench Run/Stop", p.toggleBenchCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", p.deleteCmd, true),
		ui.KeyShiftP:   ui.NewKeyAction("Sort Ports", p.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftU:   ui.NewKeyAction("Sort URL", p.GetTable().SortColCmd(4, true), false),
	})
}

func (p *PortForward) showBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if err := p.App().inject(NewBenchmark(client.NewGVR("benchmarks"))); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *PortForward) toggleBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.bench != nil {
		p.App().Status(ui.FlashErr, "Benchmark Camceled!")
		p.bench.Cancel()
		p.App().ClearStatus(true)
		return nil
	}

	sel := p.GetTable().GetSelectedItem()
	if sel == "" {
		return nil
	}

	r, _ := p.GetTable().GetSelection()
	cfg := defaultConfig()
	if b, ok := p.App().Bench.Benchmarks.Containers[sel]; ok {
		cfg = b
	}
	cfg.Name = sel

	base := ui.TrimCell(p.GetTable().SelectTable, r, 4)
	var err error
	if p.bench, err = perf.NewBenchmark(base, p.App().version, cfg); err != nil {
		p.App().Flash().Errf("Bench failed %v", err)
		p.App().ClearStatus(false)
		return nil
	}

	p.App().Status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go p.runBenchmark()

	return nil
}

func (p *PortForward) runBenchmark() {
	p.bench.Run(p.App().Config.K9s.CurrentCluster, func() {
		log.Debug().Msg("Bench Completed!")
		p.App().QueueUpdate(func() {
			if p.bench.Canceled() {
				p.App().Status(ui.FlashInfo, "Benchmark canceled")
			} else {
				p.App().Status(ui.FlashInfo, "Benchmark Completed!")
				p.bench.Cancel()
			}
			p.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				p.App().QueueUpdate(func() { p.App().ClearStatus(true) })
			}()
		})
	})
}

func (p *PortForward) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.GetTable().SearchBuff().Empty() {
		p.GetTable().SearchBuff().Reset()
		return nil
	}

	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	log.Debug().Msgf("PF DELETE %q", path)

	showModal(p.App().Content.Pages, fmt.Sprintf("Delete PortForward `%s?", path), func() {
		var pf dao.PortForward
		pf.Init(p.App().factory, client.NewGVR("portforwards"))
		if err := pf.Delete(path, true, true); err != nil {
			p.App().Flash().Err(err)
			return
		}
		p.App().Flash().Infof("PortForward %s deleted!", path)
		p.GetTable().Refresh()
	})

	return nil
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

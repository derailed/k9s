// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

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
	p.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDodgerBlue).Attributes(tcell.AttrNone))
	p.GetTable().SetSortCol(ageCol, true)
	p.SetContextFn(p.portForwardContext)
	p.AddBindKeysFn(p.bindKeys)

	return &p
}

func (p *PortForward) portForwardContext(ctx context.Context) context.Context {
	if bc := p.App().BenchFile; bc != "" {
		return context.WithValue(ctx, internal.KeyBenchCfg, p.App().BenchFile)
	}

	return ctx
}

func (p *PortForward) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("View Benchmarks", p.showBenchCmd, true),
		ui.KeyB:        ui.NewKeyAction("Benchmark Run/Stop", p.toggleBenchCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", p.deleteCmd, true),
		ui.KeyShiftP:   ui.NewKeyAction("Sort Ports", p.GetTable().SortColCmd("PORTS", true), false),
		ui.KeyShiftU:   ui.NewKeyAction("Sort URL", p.GetTable().SortColCmd("URL", true), false),
	})
}

func (p *PortForward) showBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	b := NewBenchmark(client.NewGVR("benchmarks"))
	b.SetContextFn(p.getContext)
	if err := p.App().inject(b, false); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *PortForward) getContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyDir, benchDir(p.App().Config))
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return ctx
	}

	return context.WithValue(ctx, internal.KeyPath, path)
}

func (p *PortForward) toggleBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.bench != nil {
		p.App().Status(model.FlashErr, "Benchmark Canceled!")
		p.bench.Cancel()
		p.App().ClearStatus(true)
		return nil
	}

	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	cfg := dao.BenchConfigFor(p.App().BenchFile, path)
	cfg.Name = path

	r, _ := p.GetTable().GetSelection()
	log.Debug().Msgf("PF NS %q", p.GetTable().GetModel().GetNamespace())
	col := 3
	if client.IsAllNamespaces(p.GetTable().GetModel().GetNamespace()) {
		col = 4
	}
	base := ui.TrimCell(p.GetTable().SelectTable, r, col)
	var err error
	p.bench, err = perf.NewBenchmark(base, p.App().version, cfg)
	if err != nil {
		p.App().Flash().Errf("Bench failed %v", err)
		p.App().ClearStatus(false)
		return nil
	}

	p.App().Status(model.FlashWarn, "Benchmark in progress...")
	go func() {
		if err := p.runBenchmark(); err != nil {
			log.Error().Err(err).Msgf("Benchmark run failed")
		}
	}()

	return nil
}

func (p *PortForward) runBenchmark() error {
	log.Debug().Msg("Bench starting...")

	ct, err := p.App().Config.K9s.ActiveContext()
	if err != nil {
		return err
	}
	name := p.App().Config.K9s.ActiveContextName()
	p.bench.Run(ct.ClusterName, name, func() {
		log.Debug().Msgf("Benchmark %q Completed!", name)
		p.App().QueueUpdate(func() {
			if p.bench.Canceled() {
				p.App().Status(model.FlashInfo, "Benchmark canceled")
			} else {
				p.App().Status(model.FlashInfo, "Benchmark Completed!")
				p.bench.Cancel()
			}
			p.bench = nil
			go func() {
				<-time.After(2 * time.Second)
				p.App().QueueUpdate(func() { p.App().ClearStatus(true) })
			}()
		})
	})

	return nil
}

func (p *PortForward) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.GetTable().CmdBuff().Empty() {
		p.GetTable().CmdBuff().Reset()
		return nil
	}

	selections := p.GetTable().GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}

	p.Stop()
	defer p.Start()
	var msg string
	if len(selections) > 1 {
		msg = fmt.Sprintf("Delete %d marked %s?", len(selections), p.GVR())
	} else if h, err := pfToHuman(selections[0]); err == nil {
		msg = fmt.Sprintf("Delete %s %s?", p.GVR().R(), h)
	} else {
		p.App().Flash().Err(err)
		return nil
	}

	dialog.ShowConfirm(p.App().Styles.Dialog(), p.App().Content.Pages, "Delete", msg, func() {
		for _, s := range selections {
			var pf dao.PortForward
			pf.Init(p.App().factory, client.NewGVR("portforwards"))
			if err := pf.Delete(context.Background(), s, nil, dao.DefaultGrace); err != nil {
				p.App().Flash().Err(err)
				return
			}
		}
		p.App().Flash().Infof("Successfully deleted %d PortForward!", len(selections))
		p.GetTable().Refresh()
	}, func() {})

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

var selRx = regexp.MustCompile(`\A([\w-]+)/([\w-]+)\|([\w-]+)?\|(\d+):(\d+)`)

func pfToHuman(s string) (string, error) {
	mm := selRx.FindStringSubmatch(s)
	if len(mm) < 6 {
		return "", fmt.Errorf("unable to parse selection %s", s)
	}

	return fmt.Sprintf("%s::%s %s->%s", mm[2], mm[3], mm[4], mm[5]), nil
}

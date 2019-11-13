package view

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type Service struct {
	*Resource

	bench *perf.Benchmark
}

func NewService(title, gvr string, list resource.List) ResourceViewer {
	s := Service{
		Resource: NewResource(title, gvr, list),
	}
	s.extraActionsFn = s.extraActions
	s.enterFn = s.showPods
	s.AddPage("logs", NewLogs(list.GetName(), &s), true, false)

	return &s
}

// Protocol...

func (v *Service) getList() resource.List {
	return v.list
}

func (v *Service) getSelection() string {
	return v.masterPage().GetSelectedItem()
}

func (v *Service) extraActions(aa ui.KeyActions) {
	aa[ui.KeyL] = ui.NewKeyAction("Logs", v.logsCmd, true)
	aa[tcell.KeyCtrlB] = ui.NewKeyAction("Bench", v.benchCmd, true)
	aa[tcell.KeyCtrlK] = ui.NewKeyAction("Bench Stop", v.benchStopCmd, true)
	aa[ui.KeyShiftT] = ui.NewKeyAction("Sort Type", v.sortColCmd(1, false), false)
}

func (v *Service) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}

func (v *Service) showPods(app *App, ns, res, sel string) {
	s := k8s.NewService(app.Conn())
	ns, n := namespaced(sel)
	svc, err := s.Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	if s, ok := svc.(*v1.Service); ok {
		v.showSvcPods(ns, s.Spec.Selector, v.backCmd)
	}
}

func (v *Service) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	l := v.GetPrimitive("logs").(*Logs)
	l.reload("", v, false)
	v.switchPage("logs")

	return nil
}

func (v *Service) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	if err := v.app.Config.SetActiveNamespace(v.list.GetNamespace()); err != nil {
		log.Error().Err(err).Msg("Unable to set active namespace")
	}
	v.app.inject(v)

	return nil
}

func (v *Service) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.status(ui.FlashErr, "Benchmark Canceled!")
		v.bench.Cancel()
	}
	v.app.StatusReset()

	return nil
}

func (v *Service) checkSvc(row int) error {
	svcType := trimCellRelative(v.masterPage(), row, 1)
	if svcType != "NodePort" && svcType != "LoadBalancer" {
		return errors.New("You must select a reachable service")
	}
	return nil
}

func (v *Service) getExternalPort(row int) (string, error) {
	ports := trimCellRelative(v.masterPage(), row, 5)

	pp := strings.Split(ports, " ")
	if len(pp) == 0 {
		return "", errors.New("No ports found")
	}

	// Grap the first port pair for now...
	tokens := strings.Split(pp[0], "►")
	if len(tokens) < 2 {
		return "", errors.New("No ports pair found")
	}

	return tokens[1], nil
}

func (v *Service) reloadBenchCfg() error {
	// BOZO!! Poorman Reload bench to make sure we pick up updates if any.
	path := ui.BenchConfig(v.app.Config.K9s.CurrentCluster)
	return v.app.Bench.Reload(path)
}

func (v *Service) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() || v.bench != nil {
		return evt
	}

	if err := v.reloadBenchCfg(); err != nil {
		v.app.Flash().Err(err)
		return nil
	}

	sel := v.getSelection()
	cfg, ok := v.app.Bench.Benchmarks.Services[sel]
	if !ok {
		v.app.Flash().Errf("No bench config found for service %s", sel)
		return nil
	}
	cfg.Name = sel
	log.Debug().Msgf("Benchmark config %#v", cfg)

	row, _ := v.masterPage().GetSelection()
	if err := v.checkSvc(row); err != nil {
		v.app.Flash().Err(err)
		return nil
	}
	port, err := v.getExternalPort(row)
	if err != nil {
		v.app.Flash().Err(err)
		return nil
	}
	if err := v.runBenchmark(port, cfg); err != nil {
		v.app.Flash().Errf("Benchmark failed %v", err)
		v.app.StatusReset()
		v.bench = nil
	}

	return nil
}

func (v *Service) runBenchmark(port string, cfg config.BenchConfig) error {
	var err error
	base := "http://" + cfg.HTTP.Host + ":" + port + cfg.HTTP.Path
	if v.bench, err = perf.NewBenchmark(base, cfg); err != nil {
		return err
	}

	v.app.status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go v.bench.Run(v.app.Config.K9s.CurrentCluster, v.benchDone)

	return nil
}

func (v *Service) benchDone() {
	log.Debug().Msg("Bench Completed!")
	v.app.QueueUpdate(func() {
		if v.bench.Canceled() {
			v.app.status(ui.FlashInfo, "Benchmark canceled")
		} else {
			v.app.status(ui.FlashInfo, "Benchmark Completed!")
			v.bench.Cancel()
		}
		v.bench = nil
		go benchTimedOut(v.app)
	})
}

func benchTimedOut(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.StatusReset()
	})
}

func (v *Service) showSvcPods(ns string, sel map[string]string, a ui.ActionHandler) {
	var s []string
	for k, v := range sel {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	showPods(v.app, ns, strings.Join(s, ","), "", a)
}

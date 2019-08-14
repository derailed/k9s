package views

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

type svcView struct {
	*resourceView

	bench *perf.Benchmark
}

func newSvcView(t string, app *appView, list resource.List) resourceViewer {
	v := svcView{
		resourceView: newResourceView(t, app, list).(*resourceView),
	}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods
	v.AddPage("logs", newLogsView(list.GetName(), app, &v), true, false)

	return &v
}

// Protocol...

func (v *svcView) getList() resource.List {
	return v.list
}

func (v *svcView) getSelection() string {
	return v.masterPage().GetSelectedItem()
}

func (v *svcView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyL] = ui.NewKeyAction("Logs", v.logsCmd, true)
	aa[tcell.KeyCtrlB] = ui.NewKeyAction("Bench", v.benchCmd, true)
	aa[tcell.KeyCtrlK] = ui.NewKeyAction("Bench Stop", v.benchStopCmd, true)
	aa[ui.KeyShiftT] = ui.NewKeyAction("Sort Type", v.sortColCmd(1, false), true)
}

func (v *svcView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}

func (v *svcView) showPods(app *appView, ns, res, sel string) {
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

func (v *svcView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	l := v.GetPrimitive("logs").(*logsView)
	l.reload("", v, false)
	v.switchPage("logs")

	return nil
}

func (v *svcView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.Config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}

func (v *svcView) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.status(ui.FlashErr, "Benchmark Camceled!")
		v.bench.Cancel()
	}
	v.app.StatusReset()

	return nil
}

func (v *svcView) checkSvc(row int) error {
	svcType := trimCellRelative(v.masterPage(), row, 1)
	if svcType != "NodePort" && svcType != "LoadBalancer" {
		return errors.New("You must select a reachable service")
	}
	return nil
}

func (v *svcView) getExternalPort(row int) (string, error) {
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

func (v *svcView) reloadBenchCfg() error {
	// BOZO!! Poorman Reload bench to make sure we pick up updates if any.
	path := ui.BenchConfig(v.app.Config.K9s.CurrentCluster)
	return v.app.Bench.Reload(path)
}

func (v *svcView) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
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

func (v *svcView) runBenchmark(port string, cfg config.BenchConfig) error {
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

func (v *svcView) benchDone() {
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

func benchTimedOut(app *appView) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.StatusReset()
	})
}

func (v *svcView) showSvcPods(ns string, sel map[string]string, b ui.ActionHandler) {
	var s []string
	for k, v := range sel {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	list := resource.NewPodList(v.app.Conn(), ns)
	list.SetLabelSelector(strings.Join(s, ","))

	pv := newPodView("Pods", v.app, list)
	pv.setColorerFn(podColorer)
	pv.setExtraActionsFn(func(aa ui.KeyActions) {
		aa[tcell.KeyEsc] = ui.NewKeyAction("Back", b, true)
	})
	// set active namespace to service ns.
	v.app.Config.SetActiveNamespace(ns)
	v.app.inject(pv)
}

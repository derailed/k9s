package views

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type svcView struct {
	*resourceView

	bench *benchmark
}

func newSvcView(t string, app *appView, list resource.List) resourceViewer {
	v := svcView{resourceView: newResourceView(t, app, list).(*resourceView)}
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
	return v.selectedItem
}

func (v *svcView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[tcell.KeyCtrlB] = newKeyAction("Bench", v.benchCmd, true)
	aa[KeyAltB] = newKeyAction("Bench Stop", v.benchStopCmd, true)

	aa[KeyShiftT] = newKeyAction("Sort Type", v.sortColCmd(1, false), true)
}

func (v *svcView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *svcView) showPods(app *appView, ns, res, sel string) {
	s := k8s.NewService(app.conn())
	ns, n := namespaced(sel)
	svc, err := s.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetch service %s", sel)
		app.flash().err(err)
		return
	}

	if s, ok := svc.(*v1.Service); ok {
		v.showSvcPods(ns, s.Spec.Selector, v.backCmd)
	}
}

func (v *svcView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	l := v.GetPrimitive("logs").(*logsView)
	l.reload("", v, v.list.GetName(), false)
	v.switchPage("logs")

	return nil
}

func (v *svcView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}

func (v *svcView) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.status(flashErr, "Benchmark Camceled!")
		v.bench.cancel()
	}
	v.app.statusReset()

	return nil
}

func (v *svcView) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	if v.bench != nil {
		v.app.flash().err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	sel := v.getSelectedItem()
	tv := v.getTV()
	r, _ := tv.GetSelection()

	// BOZO!! Poorman Reload bench to make sure we pick up updates if any.
	path := benchConfig(v.app.config.K9s.CurrentCluster)
	if err := v.app.bench.Reload(path); err != nil {
		log.Error().Err(err).Msg("Bench config reload")
		v.app.flash().err(err)
		return nil
	}

	cfg, ok := v.app.bench.Benchmarks.Services[sel]
	if !ok {
		v.app.flash().errf("No bench config found for service %s", sel)
		return nil
	}

	svcType := strings.TrimSpace(tv.GetCell(r, tv.nameColIndex()+1).Text)
	if svcType != "NodePort" && svcType != "LoadBalancer" {
		v.app.flash().err(errors.New("You must select a reachable service"))
		return nil
	}

	ports := strings.TrimSpace(tv.GetCell(r, tv.nameColIndex()+5).Text)
	pp := strings.Split(ports, " ")
	if len(pp) == 0 {
		v.app.flash().err(errors.New("No ports found"))
		return nil
	}
	// Grap the first port pair for now...
	tokens := strings.Split(pp[0], "►")
	if len(tokens) < 2 {
		v.app.flash().err(errors.New("No ports pair found"))
		return nil
	}
	// Found external nodeport
	port := tokens[1]
	cfg.Name = sel
	log.Debug().Msgf(">>>>> BENCHCONFIG %#v", cfg)

	base := "http://" + cfg.Host + ":" + port + cfg.Path
	var err error
	if v.bench, err = newBenchmark(base, cfg); err != nil {
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

func (v *svcView) showSvcPods(ns string, sel map[string]string, b actionHandler) {
	var s []string
	for k, v := range sel {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	list := resource.NewPodList(v.app.conn(), ns)
	list.SetLabelSelector(strings.Join(s, ","))

	pv := newPodView("Pods", v.app, list)
	pv.setColorerFn(podColorer)
	pv.setExtraActionsFn(func(aa keyActions) {
		aa[tcell.KeyEsc] = newKeyAction("Back", b, true)
	})
	// set active namespace to service ns.
	v.app.config.SetActiveNamespace(ns)
	v.app.inject(pv)
}

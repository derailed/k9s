package view

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Service represents a service viewer.
type Service struct {
	ResourceViewer

	bench *perf.Benchmark
}

// NewService returns a new viewer.
func NewService(gvr client.GVR) ResourceViewer {
	s := Service{
		ResourceViewer: NewLogsExtender(NewBrowser(gvr), nil),
	}
	s.SetBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

// Protocol...

func (s *Service) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyB:      ui.NewKeyAction("Bench", s.benchCmd, true),
		ui.KeyK:      ui.NewKeyAction("Bench Stop", s.benchStopCmd, true),
		ui.KeyShiftT: ui.NewKeyAction("Sort Type", s.GetTable().SortColCmd(1, true), false),
	})
}

func (s *Service) showPods(app *App, ns, gvr, path string) {
	o, err := app.factory.Get(gvr, path, true, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var svc v1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsWithLabels(app, path, svc.Spec.Selector)
}

func (s *Service) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		s.App().Status(ui.FlashErr, "Benchmark Canceled!")
		s.bench.Cancel()
	}
	s.App().ClearStatus(true)

	return nil
}

func (s *Service) checkSvc(row int) error {
	svcType := trimCellRelative(s.GetTable(), row, 1)
	if svcType != "NodePort" && svcType != "LoadBalancer" {
		return errors.New("You must select a reachable service")
	}
	return nil
}

func (s *Service) getExternalPort(row int) (string, error) {
	ports := trimCellRelative(s.GetTable(), row, 5)

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

func (s *Service) reloadBenchCfg() error {
	path := ui.BenchConfig(s.App().Config.K9s.CurrentCluster)
	return s.App().Bench.Reload(path)
}

func (s *Service) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := s.GetTable().GetSelectedItem()
	if sel == "" || s.bench != nil {
		return evt
	}

	if err := s.reloadBenchCfg(); err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	cfg, ok := s.App().Bench.Benchmarks.Services[sel]
	if !ok {
		s.App().Flash().Errf("No bench config found for service %s", sel)
		return nil
	}
	cfg.Name = sel
	log.Debug().Msgf("Benchmark config %#v", cfg)

	row := s.GetTable().GetSelectedRowIndex()
	if err := s.checkSvc(row); err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	port, err := s.getExternalPort(row)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	if err := s.runBenchmark(port, cfg); err != nil {
		s.App().Flash().Errf("Benchmark failed %v", err)
		s.App().ClearStatus(false)
		s.bench = nil
	}

	return nil
}

// BOZO!! Refactor used by forwards
func (s *Service) runBenchmark(port string, cfg config.BenchConfig) error {
	if cfg.HTTP.Host == "" {
		return fmt.Errorf("Invalid benchmark host %q", cfg.HTTP.Host)
	}

	var err error
	base := "http://" + cfg.HTTP.Host + ":" + port + cfg.HTTP.Path
	if s.bench, err = perf.NewBenchmark(base, s.App().version, cfg); err != nil {
		return err
	}

	s.App().Status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go s.bench.Run(s.App().Config.K9s.CurrentCluster, s.benchDone)

	return nil
}

func (s *Service) benchDone() {
	log.Debug().Msg("Bench Completed!")
	s.App().QueueUpdate(func() {
		if s.bench.Canceled() {
			s.App().Status(ui.FlashInfo, "Benchmark canceled")
		} else {
			s.App().Status(ui.FlashInfo, "Benchmark Completed!")
			s.bench.Cancel()
		}
		s.bench = nil
		go benchTimedOut(s.App())
	})
}

func benchTimedOut(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.ClearStatus(true)
	})
}

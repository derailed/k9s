package view

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Service represents a service viewer.
type Service struct {
	ResourceViewer

	bench *perf.Benchmark
}

// NewService returns a new viewer.
func NewService(gvr client.GVR) ResourceViewer {
	s := Service{
		ResourceViewer: NewPortForwardExtender(
			NewLogsExtender(NewBrowser(gvr), nil),
		),
	}
	s.SetBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

// Protocol...

func (s *Service) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyCtrlB: ui.NewKeyAction("Bench Run/Stop", s.toggleBenchCmd, true),
		ui.KeyShiftT:   ui.NewKeyAction("Sort Type", s.GetTable().SortColCmd(1, true), false),
	})
}

func (s *Service) showPods(a *App, _ ui.Tabular, gvr, path string) {
	var res dao.Service
	res.Init(a.factory, client.NewGVR(s.GVR()))

	svc, err := res.GetInstance(path)
	if err != nil {
		a.Flash().Err(err)
		return
	}

	showPodsWithLabels(a, path, svc.Spec.Selector)
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

func (s *Service) toggleBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		s.App().Status(model.FlashErr, "Benchmark Canceled!")
		s.bench.Cancel()
		s.App().ClearStatus(true)
		return nil
	}

	sel := s.GetTable().GetSelectedItem()
	if sel == "" || s.bench != nil {
		return evt
	}

	cust, err := config.NewBench(s.App().BenchFile)
	if err != nil {
		log.Debug().Msgf("No custom benchmark config file found")
	}

	cfg, ok := cust.Benchmarks.Services[sel]
	if !ok {
		s.App().Flash().Errf("No bench config found for service %s", sel)
		return nil
	}
	cfg.Name = sel
	log.Debug().Msgf("Benchmark config %#v", cfg)

	row := s.GetTable().GetSelectedRowIndex()
	if e := s.checkSvc(row); e != nil {
		s.App().Flash().Err(e)
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

	s.App().Status(model.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go s.bench.Run(s.App().Config.K9s.CurrentCluster, s.benchDone)

	return nil
}

func (s *Service) benchDone() {
	log.Debug().Msg("Bench Completed!")
	s.App().QueueUpdate(func() {
		if s.bench.Canceled() {
			s.App().Status(model.FlashInfo, "Benchmark canceled")
		} else {
			s.App().Status(model.FlashInfo, "Benchmark Completed!")
			s.bench.Cancel()
		}
		s.bench = nil
		go benchTimedOut(s.App())
	})
}

// ----------------------------------------------------------------------------
// Helpers...

func benchTimedOut(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.ClearStatus(true)
	})
}

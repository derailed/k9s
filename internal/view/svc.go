package view

import (
	"context"
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

// Service represents a service viewer.
type Service struct {
	*Resource

	bench *perf.Benchmark
	logs  *Logs
}

// NewService returns a new viewer.
func NewService(title, gvr string, list resource.List) ResourceViewer {
	return &Service{
		Resource: NewResource(title, gvr, list),
	}
}

// Init initializes the viewer.
func (s *Service) Init(ctx context.Context) {
	s.extraActionsFn = s.extraActions
	s.enterFn = s.showPods
	s.Resource.Init(ctx)

	s.logs = NewLogs(s.list.GetName(), s)
	s.logs.Init(ctx)
}

// Protocol...

func (s *Service) getList() resource.List {
	return s.list
}

func (s *Service) getSelection() string {
	return s.masterPage().GetSelectedItem()
}

func (s *Service) extraActions(aa ui.KeyActions) {
	aa[ui.KeyL] = ui.NewKeyAction("Logs", s.logsCmd, true)
	aa[tcell.KeyCtrlB] = ui.NewKeyAction("Bench", s.benchCmd, true)
	aa[tcell.KeyCtrlK] = ui.NewKeyAction("Bench Stop", s.benchStopCmd, true)
	aa[ui.KeyShiftT] = ui.NewKeyAction("Sort Type", s.sortColCmd(1, false), false)
}

func (s *Service) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	svc, err := k8s.NewService(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	if sv, ok := svc.(*v1.Service); ok {
		s.showSvcPods(ns, sv.Spec.Selector)
	}
}

func (s *Service) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.masterPage().RowSelected() {
		return evt
	}

	s.logs.reload("", s, false)
	s.Push(s.logs)

	return nil
}

func (s *Service) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		s.app.status(ui.FlashErr, "Benchmark Canceled!")
		s.bench.Cancel()
	}
	s.app.StatusReset()

	return nil
}

func (s *Service) checkSvc(row int) error {
	svcType := trimCellRelative(s.masterPage(), row, 1)
	if svcType != "NodePort" && svcType != "LoadBalancer" {
		return errors.New("You must select a reachable service")
	}
	return nil
}

func (s *Service) getExternalPort(row int) (string, error) {
	ports := trimCellRelative(s.masterPage(), row, 5)

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
	// BOZO!! Poorman Reload bench to make sure we pick up updates if any.
	path := ui.BenchConfig(s.app.Config.K9s.CurrentCluster)
	return s.app.Bench.Reload(path)
}

func (s *Service) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.masterPage().RowSelected() || s.bench != nil {
		return evt
	}

	if err := s.reloadBenchCfg(); err != nil {
		s.app.Flash().Err(err)
		return nil
	}

	sel := s.getSelection()
	cfg, ok := s.app.Bench.Benchmarks.Services[sel]
	if !ok {
		s.app.Flash().Errf("No bench config found for service %s", sel)
		return nil
	}
	cfg.Name = sel
	log.Debug().Msgf("Benchmark config %#v", cfg)

	row, _ := s.masterPage().GetSelection()
	if err := s.checkSvc(row); err != nil {
		s.app.Flash().Err(err)
		return nil
	}
	port, err := s.getExternalPort(row)
	if err != nil {
		s.app.Flash().Err(err)
		return nil
	}
	if err := s.runBenchmark(port, cfg); err != nil {
		s.app.Flash().Errf("Benchmark failed %v", err)
		s.app.StatusReset()
		s.bench = nil
	}

	return nil
}

func (s *Service) runBenchmark(port string, cfg config.BenchConfig) error {
	if cfg.HTTP.Host == "" {
		return fmt.Errorf("Invalid benchmark host %q", cfg.HTTP.Host)
	}

	var err error
	base := "http://" + cfg.HTTP.Host + ":" + port + cfg.HTTP.Path
	if s.bench, err = perf.NewBenchmark(base, cfg); err != nil {
		return err
	}

	s.app.status(ui.FlashWarn, "Benchmark in progress...")
	log.Debug().Msg("Bench starting...")
	go s.bench.Run(s.app.Config.K9s.CurrentCluster, s.benchDone)

	return nil
}

func (s *Service) benchDone() {
	log.Debug().Msg("Bench Completed!")
	s.app.QueueUpdate(func() {
		if s.bench.Canceled() {
			s.app.status(ui.FlashInfo, "Benchmark canceled")
		} else {
			s.app.status(ui.FlashInfo, "Benchmark Completed!")
			s.bench.Cancel()
		}
		s.bench = nil
		go benchTimedOut(s.app)
	})
}

func benchTimedOut(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.StatusReset()
	})
}

func (s *Service) showSvcPods(ns string, sel map[string]string) {
	var labels []string
	for k, v := range sel {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	showPods(s.app, ns, strings.Join(labels, ","), "")
}

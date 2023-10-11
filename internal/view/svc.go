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
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
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
		ResourceViewer: NewPortForwardExtender(
			NewLogsExtender(NewBrowser(gvr), nil),
		),
	}
	s.AddBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

// Protocol...

func (s *Service) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyCtrlL: ui.NewKeyAction("Bench Run/Stop", s.toggleBenchCmd, true),
		ui.KeyShiftT:   ui.NewKeyAction("Sort Type", s.GetTable().SortColCmd("TYPE", true), false),
	})
}

func (s *Service) showPods(a *App, _ ui.Tabular, gvr, path string) {
	var res dao.Service
	res.Init(a.factory, s.GVR())

	svc, err := res.GetInstance(path)
	if err != nil {
		a.Flash().Err(err)
		return
	}
	if svc.Spec.Type == v1.ServiceTypeExternalName {
		a.Flash().Warnf("No matching pods. Service %s is an external service.", path)
		return
	}

	showPodsWithLabels(a, path, svc.Spec.Selector)
}

func (s *Service) checkSvc(svc *v1.Service) error {
	if svc.Spec.Type != "NodePort" && svc.Spec.Type != "LoadBalancer" {
		return errors.New("You must select a reachable service")
	}
	return nil
}

func (s *Service) getExternalPort(svc *v1.Service) (string, error) {
	if svc.Spec.Type == "LoadBalancer" {
		return "", nil
	}
	ports := render.ToPorts(svc.Spec.Ports)
	pp := strings.Split(ports, " ")
	// Grab the first port pair for now...
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

	path := s.GetTable().GetSelectedItem()
	if path == "" || s.bench != nil {
		return evt
	}

	cust, err := config.NewBench(s.App().BenchFile)
	if err != nil {
		log.Debug().Msgf("No bench config file found %s", s.App().BenchFile)
	}

	cfg, ok := cust.Benchmarks.Services[path]
	if !ok {
		s.App().Flash().Errf("No bench config found for service %s in %s", path, s.App().BenchFile)
		return nil
	}
	cfg.Name = path
	log.Debug().Msgf("Benchmark config %#v", cfg)

	svc, err := fetchService(s.App().factory, path)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	if e := s.checkSvc(svc); e != nil {
		s.App().Flash().Err(e)
		return nil
	}
	port, err := s.getExternalPort(svc)
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

// BOZO!! Refactor used by forwards.
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
		go clearStatus(s.App())
	})
}

// ----------------------------------------------------------------------------
// Helpers...

func clearStatus(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.ClearStatus(true)
	})
}

func fetchService(f dao.Factory, path string) (*v1.Service, error) {
	o, err := f.Get("v1/services", path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var svc v1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)
	if err != nil {
		return nil, err
	}

	return &svc, nil
}

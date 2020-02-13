package view

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/perf"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
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
		ui.KeyShiftF:   ui.NewKeyAction("Port-Forward", s.portFwdCmd, true),
		tcell.KeyCtrlB: ui.NewKeyAction("Bench Run/Stop", s.toggleBenchCmd, true),
		ui.KeyShiftT:   ui.NewKeyAction("Sort Type", s.GetTable().SortColCmd(1, true), false),
	})
}

func podFromSelector(f dao.Factory, ns string, sel map[string]string) (string, error) {
	log.Debug().Msgf("Looking for pods %q:%v -- %v", ns, sel, labels.Set(sel).AsSelector())
	oo, err := f.List("v1/pods", ns, true, labels.Set(sel).AsSelector())
	if err != nil {
		return "", err
	}

	if len(oo) == 0 {
		return "", fmt.Errorf("no matching pods for %v", sel)
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(oo[0].(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return "", err
	}

	return client.FQN(pod.Namespace, pod.Name), nil
}

func (s *Service) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	svc, err := fetchService(s.App().factory, s.GVR(), path)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	ns, _ := client.Namespaced(path)
	pod, err := podFromSelector(s.App().factory, ns, svc.Spec.Selector)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	pp, err := fetchPodPorts(s.App().factory, pod)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	ports := make([]string, 0, len(pp))
	for _, p := range pp {
		if p.Protocol == v1.ProtocolTCP {
			port := fmt.Sprintf("%s:%d", p.Name, p.ContainerPort)
			if p.Name == "" {
				port = fmt.Sprintf("%d", p.ContainerPort)
			}
			ports = append(ports, port)
		}
	}

	if len(ports) == 0 {
		s.App().Flash().Err(fmt.Errorf("no tcp ports found on %s", path))
		return nil
	}

	dialog.ShowPortForwards(s.App().Content.Pages, s.App().Styles, pod, ports, s.portForward)

	return nil
}

func (s *Service) portForward(path, address, lport, cport string) {
	pf := dao.NewPortForwarder(s.App().Conn())
	ports := []string{lport + ":" + cport}
	fw, err := pf.Start(path, "", address, ports)
	if err != nil {
		s.App().Flash().Err(err)
		return
	}

	log.Debug().Msgf(">>> Starting port forward %q %v", path, ports)
	go runForward(s.App(), pf, fw)
}

func (s *Service) showPods(app *App, _ ui.Tabular, gvr, path string) {
	svc, err := fetchService(app.factory, gvr, path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsWithLabels(app, path, svc.Spec.Selector)
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

func (s *Service) toggleBenchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		s.App().Status(ui.FlashErr, "Benchmark Canceled!")
		s.bench.Cancel()
		s.App().ClearStatus(true)
		return nil
	}

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

// ----------------------------------------------------------------------------
// Helpers...

func fetchService(f dao.Factory, gvr, path string) (*v1.Service, error) {
	o, err := f.Get(gvr, path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var svc v1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)

	return &svc, err
}

func benchTimedOut(app *App) {
	<-time.After(2 * time.Second)
	app.QueueUpdate(func() {
		app.ClearStatus(true)
	})
}

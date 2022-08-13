package view

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/port"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/portforward"
)

// PortForwardExtender adds port-forward extensions.
type PortForwardExtender struct {
	ResourceViewer
}

// NewPortForwardExtender returns a new extender.
func NewPortForwardExtender(r ResourceViewer) ResourceViewer {
	p := PortForwardExtender{ResourceViewer: r}
	p.AddBindKeysFn(p.bindKeys)

	return &p
}

func (p *PortForwardExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftF: ui.NewKeyAction("Port-Forward", p.portFwdCmd, true),
	})
}

func (p *PortForwardExtender) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	podName, err := p.fetchPodName(path)
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	pod, err := fetchPod(p.App().factory, podName)
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	if pod.Status.Phase != v1.PodRunning {
		p.App().Flash().Errf("pod must be running. Current status=%v", pod.Status.Phase)
		return nil
	}
	if p.App().factory.Forwarders().IsPodForwarded(path) {
		p.App().Flash().Errf("A PortForward already exists for pod %s", pod.Name)
		return nil
	}
	if err := showFwdDialog(p, podName, startFwdCB); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *PortForwardExtender) fetchPodName(path string) (string, error) {
	res, err := dao.AccessorFor(p.App().factory, p.GVR())
	if err != nil {
		return "", err
	}
	ctrl, ok := res.(dao.Controller)
	if !ok {
		return "", fmt.Errorf("expecting a controller resource for %q", p.GVR())
	}

	return ctrl.Pod(path)
}

// ----------------------------------------------------------------------------
// Helpers...

func runForward(v ResourceViewer, pf watch.Forwarder, f *portforward.PortForwarder) {
	v.App().factory.AddForwarder(pf)

	v.App().QueueUpdateDraw(func() {
		DismissPortForwards(v, v.App().Content.Pages)
	})

	pf.SetActive(true)
	if err := f.ForwardPorts(); err != nil {
		v.App().Flash().Err(err)
		return
	}

	v.App().QueueUpdateDraw(func() {
		v.App().factory.DeleteForwarder(pf.FQN())
		pf.SetActive(false)
	})
}

func startFwdCB(v ResourceViewer, path string, pts port.PortTunnels) error {
	if err := pts.CheckAvailable(); err != nil {
		return err
	}

	tt := make([]string, 0, len(pts))
	for _, pt := range pts {
		if _, ok := v.App().factory.ForwarderFor(dao.PortForwardID(path, pt.Container, pt.PortMap())); ok {
			return fmt.Errorf("A port-forward is already active on pod %s", path)
		}
		pf := dao.NewPortForwarder(v.App().factory)
		fwd, err := pf.Start(path, pt)
		if err != nil {
			return err
		}
		log.Debug().Msgf(">>> Starting port forward %q -- %#v", pf.ID(), pt)
		go runForward(v, pf, fwd)
		tt = append(tt, pt.ContainerPort)
	}
	if len(tt) == 1 {
		v.App().Flash().Infof("PortForward activated %s", tt[0])
		return nil
	}
	v.App().Flash().Infof("PortForwards activated %s", strings.Join(tt, ","))

	return nil
}

func showFwdDialog(v ResourceViewer, path string, cb PortForwardCB) error {
	mm, anns, err := fetchPodPorts(v.App().factory, path)
	if err != nil {
		return err
	}
	ports := make(port.ContainerPortSpecs, 0, len(mm))
	for co, pp := range mm {
		for _, p := range pp {
			if p.Protocol != v1.ProtocolTCP {
				continue
			}
			ports = append(ports, port.NewPortSpec(co, p.Name, p.ContainerPort))
		}
	}
	if spec, ok := anns[port.K9sAutoPortForwardsKey]; ok {
		pfs, err := port.ParsePFs(spec)
		if err != nil {
			return err
		}

		pts, err := pfs.ToTunnels(v.App().Config.CurrentCluster().PortForwardAddress, ports, port.IsPortFree)
		if err != nil {
			return err
		}

		return startFwdCB(v, path, pts)
	}

	ShowPortForwards(v, path, ports, anns, cb)

	return nil
}

func fetchPodPorts(f *watch.Factory, path string) (map[string][]v1.ContainerPort, map[string]string, error) {
	log.Debug().Msgf("Fetching ports on pod %q", path)
	o, err := f.Get("v1/pods", path, true, labels.Everything())
	if err != nil {
		return nil, nil, err
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return nil, nil, err
	}

	pp := make(map[string][]v1.ContainerPort, len(pod.Spec.Containers))
	for _, co := range pod.Spec.Containers {
		pp[co.Name] = co.Ports
	}

	return pp, pod.Annotations, nil
}

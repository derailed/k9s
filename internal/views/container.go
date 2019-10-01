package views

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

type containerView struct {
	*logResourceView

	current ui.Igniter
	exitFn  func()
}

func newContainerView(title string, app *appView, list resource.List, path string, exitFn func()) resourceViewer {
	v := containerView{logResourceView: newLogResourceView(title, "", app, list)}
	v.path = &path
	v.envFn = v.k9sEnv
	v.containerFn = v.selectedContainer
	v.extraActionsFn = v.extraActions
	v.enterFn = v.viewLogs
	v.colorerFn = containerColorer
	v.current = app.Frame().GetPrimitive("main").(ui.Igniter)
	v.exitFn = exitFn

	return &v
}

func (v *containerView) Init(ctx context.Context, ns string) {
	v.resourceView.Init(ctx, ns)
}

func (v *containerView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)

	aa[ui.KeyShiftF] = ui.NewKeyAction("PortForward", v.portFwdCmd, true)
	aa[ui.KeyShiftL] = ui.NewKeyAction("Logs Previous", v.prevLogsCmd, true)
	aa[ui.KeyS] = ui.NewKeyAction("Shell", v.shellCmd, true)
	aa[tcell.KeyEscape] = ui.NewKeyAction("Back", v.backCmd, false)
	aa[ui.KeyP] = ui.NewKeyAction("Previous", v.backCmd, false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", v.sortColCmd(6, false), false)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", v.sortColCmd(7, false), false)
	aa[ui.KeyShiftX] = ui.NewKeyAction("Sort CPU%", v.sortColCmd(8, false), false)
	aa[ui.KeyShiftZ] = ui.NewKeyAction("Sort MEM%", v.sortColCmd(9, false), false)
}

func (v *containerView) k9sEnv() K9sEnv {
	env := v.defaultK9sEnv()

	ns, n := namespaced(*v.path)
	env["POD"] = n
	env["NAMESPACE"] = ns
	log.Debug().Msgf("OVER ENV %#v", env)

	return env
}

func (v *containerView) selectedContainer() string {
	return v.masterPage().GetSelectedItem()
}

func (v *containerView) viewLogs(app *appView, _, res, sel string) {
	status := v.masterPage().GetSelectedCell(3)
	if status == "Running" || status == "Completed" {
		v.showLogs(false)
		return
	}
	v.app.Flash().Err(errors.New("No logs available"))
}

// Handlers...

func (v *containerView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	v.stopUpdates()
	shellIn(v.app, *v.path, v.masterPage().GetSelectedItem())
	v.restartUpdates()
	return nil
}

func (v *containerView) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItem()
	if _, ok := v.app.forwarders[fwFQN(*v.path, sel)]; ok {
		v.app.Flash().Err(fmt.Errorf("A PortForward already exist on container %s", *v.path))
		return nil
	}

	state := v.masterPage().GetSelectedCell(3)
	if state != "Running" {
		v.app.Flash().Err(fmt.Errorf("Container %s is not running?", sel))
		return nil
	}

	portC := v.masterPage().GetSelectedCell(10)
	ports := strings.Split(portC, ",")
	if len(ports) == 0 {
		v.app.Flash().Err(errors.New("Container exposes no ports"))
		return nil
	}

	var port string
	for _, p := range ports {
		log.Debug().Msgf("Checking port %q", p)
		if !isTCPPort(p) {
			continue
		}
		port = strings.TrimSpace(p)
		break
	}
	if port == "" {
		v.app.Flash().Warn("No valid TCP port found on this container. User will specify...")
		port = "MY_TCP_PORT!"
	}
	dialog.ShowPortForward(v.Pages, port, v.portForward)

	return nil
}

func (v *containerView) portForward(lport, cport string) {
	co := v.masterPage().GetSelectedCell(0)
	pf := k8s.NewPortForward(v.app.Conn(), &log.Logger)
	ports := []string{lport + ":" + cport}
	fw, err := pf.Start(*v.path, co, ports)
	if err != nil {
		v.app.Flash().Err(err)
		return
	}

	log.Debug().Msgf(">>> Starting port forward %q %v", *v.path, ports)
	go v.runForward(pf, fw)
}

func (v *containerView) runForward(pf *k8s.PortForward, f *portforward.PortForwarder) {
	v.app.QueueUpdateDraw(func() {
		v.app.forwarders[pf.FQN()] = pf
		v.app.Flash().Infof("PortForward activated %s:%s", pf.Path(), pf.Ports()[0])
		dialog.DismissPortForward(v.Pages)
	})

	pf.SetActive(true)
	if err := f.ForwardPorts(); err != nil {
		v.app.Flash().Err(err)
		return
	}
	v.app.QueueUpdateDraw(func() {
		delete(v.app.forwarders, pf.FQN())
		pf.SetActive(false)
	})
}

func (v *containerView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.exitFn()
	return nil
}

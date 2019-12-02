package view

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

const containerTitle = "Containers"

// Container represents a container view.
type Container struct {
	ResourceViewer

	podPath string
}

// New Container returns a new container view.
func NewContainer(path string, list resource.List) ResourceViewer {
	return &Container{
		ResourceViewer: NewResource(containerTitle, "containers", list),
		podPath:        path,
	}
}

// Init initializes the viewer.
func (c *Container) Init(ctx context.Context) error {
	c.ResourceViewer = NewLogsExtender(c.ResourceViewer, c.selectedContainer)
	c.ResourceViewer.SetPath(c.podPath)
	c.GetTable().Path = c.podPath
	if err := c.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	c.SetEnvFn(c.k9sEnv)
	c.GetTable().SetEnterFn(c.viewLogs)
	c.GetTable().SetColorerFn(render.Container{}.ColorerFunc())
	c.bindKeys()

	return nil
}

// Name returns the component name.
func (c *Container) Name() string { return containerTitle }

func (c *Container) bindKeys() {
	c.Actions().Delete(tcell.KeyCtrlSpace, ui.KeySpace)
	c.Actions().Add(ui.KeyActions{
		tcell.KeyCtrlF: ui.NewKeyAction("PortForward", c.portFwdCmd, true),
		ui.KeyS:        ui.NewKeyAction("Shell", c.shellCmd, true),
		ui.KeyShiftC:   ui.NewKeyAction("Sort CPU", c.GetTable().SortColCmd(6, false), false),
		ui.KeyShiftM:   ui.NewKeyAction("Sort MEM", c.GetTable().SortColCmd(7, false), false),
		ui.KeyShiftX:   ui.NewKeyAction("Sort CPU%", c.GetTable().SortColCmd(8, false), false),
		ui.KeyShiftZ:   ui.NewKeyAction("Sort MEM%", c.GetTable().SortColCmd(9, false), false),
	})
}

func (c *Container) k9sEnv() K9sEnv {
	env := defaultK9sEnv(c.App(), c.GetTable().GetSelectedItem(), c.GetTable().GetRow())
	ns, n := namespaced(c.podPath)
	env["POD"] = n
	env["NAMESPACE"] = ns

	return env
}

func (c *Container) selectedContainer() string {
	log.Debug().Msgf("Container SELECTED %s", c.GetTable().GetSelectedItem())
	tokens := strings.Split(c.GetTable().GetSelectedItem(), "/")

	return tokens[0]
}

func (c *Container) viewLogs(_ *App, ns, res, path string) {
	log.Debug().Msgf(">>>>>>>> ViewLOgs %q -- %q -- %q", ns, res, path)
	status := c.GetTable().GetSelectedCell(3)
	if status != "Running" && status != "Completed" {
		c.App().Flash().Err(errors.New("No logs available"))
		return
	}
	c.ResourceViewer.(*LogsExtender).showLogs(c.podPath, false)
}

// Handlers...

func (c *Container) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	c.Stop()
	defer c.Start()
	shellIn(c.App(), c.podPath, sel)

	return nil
}

func (c *Container) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	if _, ok := c.App().forwarders[fwFQN(c.podPath, sel)]; ok {
		c.App().Flash().Err(fmt.Errorf("A PortForward already exist on container %s", c.podPath))
		return nil
	}

	state := c.GetTable().GetSelectedCell(3)
	if state != "Running" {
		c.App().Flash().Err(fmt.Errorf("Container %s is not running?", sel))
		return nil
	}

	portC := c.GetTable().GetSelectedCell(11)
	ports := strings.Split(portC, ",")
	if len(ports) == 0 {
		c.App().Flash().Err(errors.New("Container exposes no ports"))
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
		c.App().Flash().Warn("No valid TCP port found on this container. User will specify...")
		port = "MY_TCP_PORT!"
	}
	dialog.ShowPortForward(c.App().Content.Pages, port, c.portForward)

	return nil
}

func (c *Container) portForward(lport, cport string) {
	co := c.GetTable().GetSelectedCell(0)
	pf := k8s.NewPortForward(c.App().Conn(), &log.Logger)
	ports := []string{lport + ":" + cport}
	fw, err := pf.Start(c.podPath, co, ports)
	if err != nil {
		c.App().Flash().Err(err)
		return
	}

	log.Debug().Msgf(">>> Starting port forward %q %v", c.podPath, ports)
	go c.runForward(pf, fw)
}

func (c *Container) runForward(pf *k8s.PortForward, f *portforward.PortForwarder) {
	c.App().QueueUpdateDraw(func() {
		c.App().forwarders[pf.FQN()] = pf
		c.App().Flash().Infof("PortForward activated %s:%s", pf.Path(), pf.Ports()[0])
		dialog.DismissPortForward(c.App().Content.Pages)
	})

	pf.SetActive(true)
	if err := f.ForwardPorts(); err != nil {
		c.App().Flash().Err(err)
		return
	}
	c.App().QueueUpdateDraw(func() {
		delete(c.App().forwarders, pf.FQN())
		pf.SetActive(false)
	})
}

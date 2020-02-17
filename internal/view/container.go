package view

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

const (
	containerTitle = "Containers"
	portsCol       = 13
)

// Container represents a container view.
type Container struct {
	ResourceViewer
}

// NewContainer returns a new container view.
func NewContainer(gvr client.GVR) ResourceViewer {
	c := Container{}
	c.ResourceViewer = NewLogsExtender(NewBrowser(gvr), c.selectedContainer)
	c.SetEnvFn(c.k9sEnv)
	c.GetTable().SetEnterFn(c.viewLogs)
	c.GetTable().SetColorerFn(render.Container{}.ColorerFunc())
	c.SetBindKeysFn(c.bindKeys)

	return &c
}

// Name returns the component name.
func (c *Container) Name() string { return containerTitle }

func (c *Container) bindDangerousKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyS: ui.NewKeyAction("Shell", c.shellCmd, true),
		ui.KeyA: ui.NewKeyAction("Attach", c.attachCmd, true),
	})
}

func (c *Container) bindKeys(aa ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlSpace, ui.KeySpace)

	if !c.App().Config.K9s.GetReadOnly() {
		c.bindDangerousKeys(aa)
	}

	aa.Add(ui.KeyActions{
		ui.KeyShiftF:   ui.NewKeyAction("PortForward", c.portFwdCmd, true),
		ui.KeyShiftC:   ui.NewKeyAction("Sort CPU", c.GetTable().SortColCmd(6, false), false),
		ui.KeyShiftM:   ui.NewKeyAction("Sort MEM", c.GetTable().SortColCmd(7, false), false),
		ui.KeyShiftX:   ui.NewKeyAction("Sort %CPU (REQ)", c.GetTable().SortColCmd(8, false), false),
		ui.KeyShiftZ:   ui.NewKeyAction("Sort %MEM (REQ)", c.GetTable().SortColCmd(9, false), false),
		tcell.KeyCtrlX: ui.NewKeyAction("Sort %CPU (LIM)", c.GetTable().SortColCmd(8, false), false),
		tcell.KeyCtrlZ: ui.NewKeyAction("Sort %MEM (LIM)", c.GetTable().SortColCmd(9, false), false),
	})
}

func (c *Container) k9sEnv() K9sEnv {
	env := defaultK9sEnv(c.App(), c.GetTable().GetSelectedItem(), c.GetTable().GetSelectedRow())
	ns, n := client.Namespaced(c.GetTable().Path)
	env["POD"] = n
	env["NAMESPACE"] = ns

	return env
}

func (c *Container) selectedContainer() string {
	tokens := strings.Split(c.GetTable().GetSelectedItem(), "/")
	return tokens[0]
}

func (c *Container) viewLogs(app *App, model ui.Tabular, gvr, path string) {
	c.ResourceViewer.(*LogsExtender).showLogs(c.GetTable().Path, false)
}

// Handlers...

func (c *Container) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	c.Stop()
	defer c.Start()
	shellIn(c.App(), c.GetTable().Path, sel)

	return nil
}

func (c *Container) attachCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	c.Stop()
	defer c.Start()
	attachIn(c.App(), c.GetTable().Path)

	return nil
}

func (c *Container) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if _, ok := c.App().factory.ForwarderFor(fwFQN(c.GetTable().Path, path)); ok {
		c.App().Flash().Err(fmt.Errorf("A PortForward already exist on container %s", c.GetTable().Path))
		return nil
	}

	ports, ok := c.isForwardable(path)
	if !ok {
		return nil
	}

	dialog.ShowPortForward(c.App().Content.Pages, c.preparePort(ports), c.portForward)

	return nil
}

func (c *Container) isForwardable(path string) ([]string, bool) {
	state := c.GetTable().GetSelectedCell(3)
	if state != "Running" {
		c.App().Flash().Err(fmt.Errorf("Container %s is not running?", path))
		return nil, false
	}

	portC := c.GetTable().GetSelectedCell(portsCol)
	ports := strings.Split(portC, ",")
	if len(ports) == 0 {
		c.App().Flash().Err(errors.New("Container exposes no ports"))
		return nil, false
	}

	return ports, true
}

func (c *Container) preparePort(pp []string) string {
	var port string
	for _, p := range pp {
		if !isTCPPort(p) {
			continue
		}
		port = strings.TrimSpace(p)
		tokens := strings.Split(port, ":")
		if len(tokens) == 2 {
			port = tokens[1]
		}
		break
	}
	if port == "" {
		c.App().Flash().Warn("No valid TCP port found on this container. User will specify...")
		return "MY_TCP_PORT!"
	}

	return port
}

func (c *Container) portForward(address, lport, cport string) {
	co := c.GetTable().GetSelectedCell(0)
	pf := dao.NewPortForwarder(c.App().Conn())
	ports := []string{lport + ":" + cport}
	fw, err := pf.Start(c.GetTable().Path, co, address, ports)
	if err != nil {
		c.App().Flash().Err(err)
		return
	}

	log.Debug().Msgf(">>> Starting port forward %q %v", c.GetTable().Path, ports)
	go c.runForward(pf, fw)
}

func (c *Container) runForward(pf *dao.PortForwarder, f *portforward.PortForwarder) {
	c.App().QueueUpdateDraw(func() {
		c.App().factory.AddForwarder(pf)
		c.App().Flash().Infof("PortForward activated %s:%s", pf.Path(), pf.Ports()[0])
		dialog.DismissPortForward(c.App().Content.Pages)
	})

	pf.SetActive(true)
	if err := f.ForwardPorts(); err != nil {
		c.App().Flash().Err(err)
		return
	}
	c.App().QueueUpdateDraw(func() {
		c.App().factory.DeleteForwarder(pf.FQN())
		pf.SetActive(false)
	})
}

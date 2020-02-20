package view

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
		tcell.KeyCtrlQ: ui.NewKeyAction("Sort %MEM (LIM)", c.GetTable().SortColCmd(9, false), false),
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

func (c *Container) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	log.Debug().Msgf("CONTAINER-SEL %q", path)
	if _, ok := c.App().factory.ForwarderFor(fwFQN(c.GetTable().Path, path)); ok {
		c.App().Flash().Err(fmt.Errorf("A PortForward already exist on container %s", c.GetTable().Path))
		return nil
	}

	ports, ok := c.isForwardable(path)
	if !ok {
		return nil
	}
	log.Debug().Msgf("CONTAINER-PORTS %#v", ports)
	ShowPortForwards(c, c.GetTable().Path, ports, startFwdCB)

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

	pp := make([]string, 0, len(ports))
	for _, p := range ports {
		if !isTCPPort(p) {
			continue
		}
		pp = append(pp, path+"/"+p)
	}

	return pp, true
}

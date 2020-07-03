package view

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const containerTitle = "Containers"

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
	c.GetTable().SetDecorateFn(c.decorateRows)
	c.SetBindKeysFn(c.bindKeys)
	c.GetTable().SetDecorateFn(c.portForwardIndicator)

	return &c
}

func (c *Container) portForwardIndicator(data render.TableData) render.TableData {
	ff := c.App().factory.Forwarders()

	col := data.IndexOfHeader("PF")
	for _, re := range data.RowEvents {
		if ff.IsContainerForwarded(c.GetTable().Path, re.Row.ID) {
			re.Row.Fields[col] = "[orange::b]Ⓕ"
		}
	}

	return data
}

func (c *Container) decorateRows(data render.TableData) render.TableData {
	return decorateCpuMemHeaderRows(c.App(), data)
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
		ui.KeyF:      ui.NewKeyAction("Show PortForward", c.showPFCmd, true),
		ui.KeyShiftF: ui.NewKeyAction("PortForward", c.portFwdCmd, true),
		ui.KeyShiftT: ui.NewKeyAction("Sort Restart", c.GetTable().SortColCmd("RESTARTS", false), false),
	})
	aa.Add(resourceSorters(c.GetTable()))
}

func (c *Container) k9sEnv() Env {
	path := c.GetTable().GetSelectedItem()
	row, ok := c.GetTable().GetSelectedRow(path)
	if !ok {
		log.Error().Msgf("unable to locate selected row for %q", path)
	}
	env := defaultEnv(c.App().Conn().Config(), path, c.GetTable().GetModel().Peek().Header, row)
	env["NAMESPACE"], env["POD"] = client.Namespaced(c.GetTable().Path)

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

func (c *Container) showPFCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if !c.App().factory.Forwarders().IsContainerForwarded(c.GetTable().Path, path) {
		c.App().Flash().Errf("no portforwards defined")
		return nil
	}
	pf := NewPortForward(client.NewGVR("portforwards"))
	pf.SetContextFn(c.portForwardContext)
	if err := c.App().inject(pf); err != nil {
		c.App().Flash().Err(err)
	}

	return nil
}

func (c *Container) portForwardContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyBenchCfg, c.App().BenchFile)
	return context.WithValue(ctx, internal.KeyPath, c.GetTable().Path)
}

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
	attachIn(c.App(), c.GetTable().Path, sel)

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
	ShowPortForwards(c, c.GetTable().Path, ports, startFwdCB)

	return nil
}

func (c *Container) isForwardable(path string) ([]string, bool) {
	po, err := fetchPod(c.App().factory, c.GetTable().Path)
	if err != nil {
		return nil, false
	}

	cc := po.Spec.Containers
	var co *v1.Container
	for i := range cc {
		if cc[i].Name == path {
			co = &cc[i]
		}
	}
	if co == nil {
		log.Error().Err(fmt.Errorf("unable to locate container named %q", path))
		return nil, false
	}

	var cs *v1.ContainerStatus
	ss := po.Status.ContainerStatuses
	for i := range ss {
		if ss[i].Name == path {
			cs = &ss[i]
		}
	}
	if cs == nil {
		log.Error().Err(fmt.Errorf("unable to locate container status for %q", path))
		return nil, false
	}

	if render.ToContainerState(cs.State) != "Running" {
		c.App().Flash().Err(fmt.Errorf("Container %s is not running?", path))
		return nil, false
	}

	portC := render.ToContainerPorts(co.Ports)
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

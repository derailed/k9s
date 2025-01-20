// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/rs/zerolog/log"
)

var (
	customViewers MetaViewers
	contextRX     = regexp.MustCompile(`\s+@([\w-]+)`)
)

// Command represents a user command.
type Command struct {
	app   *App
	alias *dao.Alias
	mx    sync.Mutex
}

// NewCommand returns a new command.
func NewCommand(app *App) *Command {
	return &Command{
		app: app,
	}
}

// AliasesFor gather all known aliases for a given resource.
func (c *Command) AliasesFor(s string) []string {
	return c.alias.AliasesFor(s)
}

// Init initializes the command.
func (c *Command) Init(path string) error {
	c.alias = dao.NewAlias(c.app.factory)
	if _, err := c.alias.Ensure(path); err != nil {
		log.Error().Err(err).Msgf("Alias ensure failed!")
		return err
	}
	customViewers = loadCustomViewers()

	return nil
}

// Reset resets Command and reload aliases.
func (c *Command) Reset(path string, clear bool) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if clear {
		c.alias.Clear()
	}
	if _, err := c.alias.Ensure(path); err != nil {
		return err
	}

	return nil
}

func allowedXRay(gvr client.GVR) bool {
	gg := map[string]struct{}{
		"v1/pods":              {},
		"v1/services":          {},
		"apps/v1/deployments":  {},
		"apps/v1/daemonsets":   {},
		"apps/v1/statefulsets": {},
		"apps/v1/replicasets":  {},
	}
	_, ok := gg[gvr.String()]

	return ok
}

func (c *Command) contextCmd(p *cmd.Interpreter) error {
	ct, ok := p.ContextArg()
	if !ok {
		return fmt.Errorf("invalid command use `context xxx`")
	}

	if ct != "" {
		return useContext(c.app, ct)
	}

	gvr, v, err := c.viewMetaFor(p)
	if err != nil {
		return err
	}

	return c.exec(p, gvr, c.componentFor(gvr, ct, v), true)
}

func (c *Command) namespaceCmd(p *cmd.Interpreter) bool {
	ns, ok := p.NSArg()
	if !ok {
		return false
	}

	if ns != "" {
		_ = p.Reset("pod " + ns)
	}

	return false
}

func (c *Command) aliasCmd(p *cmd.Interpreter) error {
	filter, _ := p.FilterArg()

	gvr := client.NewGVR("aliases")
	v := NewAlias(gvr)
	v.SetFilter(filter)

	return c.exec(p, gvr, v, false)
}

func (c *Command) xrayCmd(p *cmd.Interpreter) error {
	arg, cns, ok := p.XrayArgs()
	if !ok {
		return errors.New("invalid command. use `xray xxx`")
	}
	gvr, _, ok := c.alias.AsGVR(arg)
	if !ok {
		return fmt.Errorf("invalid resource name: %q", arg)
	}
	if !allowedXRay(gvr) {
		return fmt.Errorf("unsupported resource %q", arg)
	}
	ns := c.app.Config.ActiveNamespace()
	if cns != "" {
		ns = cns
	}
	if err := c.app.Config.SetActiveNamespace(client.CleanseNamespace(ns)); err != nil {
		return err
	}
	if err := c.app.switchNS(ns); err != nil {
		return err
	}

	return c.exec(p, client.NewGVR("xrays"), NewXray(gvr), true)
}

// Run execs the command by showing associated display.
func (c *Command) run(p *cmd.Interpreter, fqn string, clearStack bool) error {
	if c.specialCmd(p) {
		return nil
	}
	gvr, v, err := c.viewMetaFor(p)
	if err != nil {
		return err
	}

	if context, ok := p.HasContext(); ok {
		if context != c.app.Config.ActiveContextName() {
			if err := c.app.Config.Save(true); err != nil {
				log.Error().Err(err).Msg("config save failed!")
			} else {
				log.Debug().Msgf("Saved context config for: %q", context)
			}
		}
		res, err := dao.AccessorFor(c.app.factory, client.NewGVR("contexts"))
		if err != nil {
			return err
		}
		switcher, ok := res.(dao.Switchable)
		if !ok {
			return errors.New("expecting a switchable resource")
		}
		if err := switcher.Switch(context); err != nil {
			log.Error().Err(err).Msgf("Context switch failed")
			return err
		}
		if err := c.app.switchContext(p, false); err != nil {
			return err
		}
	}

	ns := c.app.Config.ActiveNamespace()
	if cns, ok := p.NSArg(); ok {
		ns = cns
	}
	if err := c.app.switchNS(ns); err != nil {
		return err
	}

	co := c.componentFor(gvr, fqn, v)
	co.SetFilter("")
	co.SetLabelFilter(nil)
	if f, ok := p.FilterArg(); ok {
		co.SetFilter(f)
	}
	if f, ok := p.FuzzyArg(); ok {
		co.SetFilter("-f " + f)
	}
	if ll, ok := p.LabelsArg(); ok {
		co.SetLabelFilter(ll)
	}

	return c.exec(p, gvr, co, clearStack)
}

func (c *Command) defaultCmd(isRoot bool) error {
	if c.app.Conn() == nil || !c.app.Conn().ConnectionOK() {
		return c.run(cmd.NewInterpreter("context"), "", true)
	}

	defCmd := "pod"
	if isRoot {
		defCmd = "ctx"
	}
	p := cmd.NewInterpreter(c.app.Config.ActiveView())
	if p.IsBlank() {
		return c.run(p.Reset(defCmd), "", true)
	}

	if err := c.run(p, "", true); err != nil {
		p = p.Reset(defCmd)
		log.Error().Err(fmt.Errorf("Command failed. Using default command: %s", p.GetLine()))
		return c.run(p, "", true)
	}

	return nil
}

func (c *Command) specialCmd(p *cmd.Interpreter) bool {
	switch {
	case p.IsCowCmd():
		if msg, ok := p.CowArg(); !ok {
			c.app.Flash().Errf("Invalid command. Use `cow xxx`")
		} else {
			c.app.cowCmd(msg)
		}
	case p.IsBailCmd():
		c.app.BailOut()
	case p.IsHelpCmd():
		_ = c.app.helpCmd(nil)
	case p.IsAliasCmd():
		if err := c.aliasCmd(p); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsXrayCmd():
		if err := c.xrayCmd(p); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsRBACCmd():
		if cat, sub, ok := p.RBACArgs(); !ok {
			c.app.Flash().Errf("Invalid command. Use `can [u|g|s]:xxx`")
		} else if err := c.app.inject(NewPolicy(c.app, cat, sub), true); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsContextCmd():
		if err := c.contextCmd(p); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsNamespaceCmd():
		return c.namespaceCmd(p)
	case p.IsDirCmd():
		if a, ok := p.DirArg(); !ok {
			c.app.Flash().Errf("Invalid command. Use `dir xxx`")
		} else if err := c.app.dirCmd(a); err != nil {
			c.app.Flash().Err(err)
		}
	default:
		return false
	}

	return true
}

func (c *Command) viewMetaFor(p *cmd.Interpreter) (client.GVR, *MetaViewer, error) {
	agvr, exp, ok := c.alias.AsGVR(p.Cmd())
	if !ok {
		return client.NoGVR, nil, fmt.Errorf("`%s` command not found", p.Cmd())
	}
	gvr := agvr
	if exp != "" {
		ff := strings.Fields(exp)
		ff[0] = agvr.String()
		ap := cmd.NewInterpreter(strings.Join(ff, " "))
		gvr = client.NewGVR(ap.Cmd())
		p.Amend(ap)
	}

	v := MetaViewer{
		viewerFn: func(gvr client.GVR) ResourceViewer {
			return NewOwnerExtender(NewBrowser(gvr))
		},
	}
	if mv, ok := customViewers[gvr]; ok {
		v = mv
	}

	return gvr, &v, nil
}

func (c *Command) componentFor(gvr client.GVR, fqn string, v *MetaViewer) ResourceViewer {
	var view ResourceViewer
	if v.viewerFn != nil {
		view = v.viewerFn(gvr)
	} else {
		view = NewBrowser(gvr)
	}

	view.SetInstance(fqn)
	if v.enterFn != nil {
		view.GetTable().SetEnterFn(v.enterFn)
	}

	return view
}

func (c *Command) exec(p *cmd.Interpreter, gvr client.GVR, comp model.Component, clearStack bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error().Msgf("Something bad happened! %#v", e)
			c.app.Content.Dump()
			log.Debug().Msgf("History %v", c.app.cmdHistory.List())
			log.Error().Msg(string(debug.Stack()))

			p := cmd.NewInterpreter("pod")
			if cmd := c.app.cmdHistory.Pop(); cmd != "" {
				p = p.Reset(cmd)
			}
			err = c.run(p, "", true)
		}
	}()

	if comp == nil {
		return fmt.Errorf("no component found for %s", gvr)
	}
	c.app.Flash().Infof("Viewing %s...", gvr.R())
	if clearStack {
		cmd := contextRX.ReplaceAllString(p.GetLine(), "")
		c.app.Config.SetActiveView(cmd)
	}
	if err := c.app.inject(comp, clearStack); err != nil {
		return err
	}

	c.app.cmdHistory.Push(p.GetLine())

	return
}

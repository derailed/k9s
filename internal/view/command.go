// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/view/cmd"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	podCmd = "v1/pods"
	ctxCmd = "ctx"
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
func (c *Command) AliasesFor(gvr *client.GVR) sets.Set[string] {
	return c.alias.AliasesFor(gvr)
}

// Init initializes the command.
func (c *Command) Init(path string) error {
	c.alias = dao.NewAlias(c.app.factory)
	if _, err := c.alias.Ensure(path); err != nil {
		slog.Error("Ensure aliases failed", slogs.Error, err)
		return err
	}
	customViewers = loadCustomViewers()

	return nil
}

// Reset resets Command and reload aliases.
func (c *Command) Reset(path string, nuke bool) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if nuke {
		c.alias.Clear()
	}
	if _, err := c.alias.Ensure(path); err != nil {
		return err
	}

	return nil
}

var allowedCmds = sets.New[*client.GVR](
	client.PodGVR,
	client.SvcGVR,
	client.DpGVR,
	client.DsGVR,
	client.StsGVR,
	client.RsGVR,
)

func allowedXRay(gvr *client.GVR) bool {
	return allowedCmds.Has(gvr)
}

func (c *Command) contextCmd(p *cmd.Interpreter, pushCmd bool) error {
	ct, ok := p.ContextArg()
	if !ok {
		return fmt.Errorf("invalid command use `context xxx`")
	}

	if ct != "" {
		return useContext(c.app, ct)
	}

	gvr, v, comd, err := c.viewMetaFor(p)
	if err != nil {
		return err
	}
	if comd != nil {
		p = comd
	}

	return c.exec(p, gvr, c.componentFor(gvr, ct, v), true, pushCmd)
}

func (*Command) namespaceCmd(p *cmd.Interpreter) bool {
	ns, ok := p.NSArg()
	if !ok {
		return false
	}

	if ns != "" {
		_ = p.Reset(client.PodGVR.String())
		p.SwitchNS(ns)
	}

	return false
}

func (c *Command) aliasCmd(p *cmd.Interpreter, pushCmd bool) error {
	filter, _ := p.FilterArg()

	v := NewAlias(client.AliGVR)
	v.SetFilter(filter, true)

	return c.exec(p, client.AliGVR, v, false, pushCmd)
}

func (c *Command) xrayCmd(p *cmd.Interpreter, pushCmd bool) error {
	arg, cns, ok := p.XrayArgs()
	if !ok {
		return errors.New("invalid command. use `xray xxx`")
	}
	gvr, ok := c.alias.Resolve(p)
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

	return c.exec(p, client.XGVR, NewXray(gvr), true, pushCmd)
}

// Run execs the command by showing associated display.
func (c *Command) run(p *cmd.Interpreter, fqn string, clearStack, pushCmd bool) error {
	if c.specialCmd(p, pushCmd) {
		return nil
	}
	gvr, v, comd, err := c.viewMetaFor(p)
	if err != nil {
		return err
	}
	if comd != nil {
		p.Merge(comd)
	}

	if context, ok := p.HasContext(); ok {
		if context != c.app.Config.ActiveContextName() {
			if err := c.app.Config.Save(true); err != nil {
				slog.Error("Config save failed during command exec", slogs.Error, err)
			} else {
				slog.Debug("Successfully saved config", slogs.Context, context)
			}
		}
		res, err := dao.AccessorFor(c.app.factory, client.CtGVR)
		if err != nil {
			return err
		}
		switcher, ok := res.(dao.Switchable)
		if !ok {
			return errors.New("expecting a switchable resource")
		}
		if err := switcher.Switch(context); err != nil {
			slog.Error("Unable to switch context", slogs.Error, err)
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
	if ok, err := dao.MetaAccess.IsNamespaced(gvr); ok && err == nil {
		if err := c.app.switchNS(ns); err != nil {
			return err
		}
		p.SwitchNS(ns)
	} else {
		p.ClearNS()
	}

	co := c.componentFor(gvr, fqn, v)
	co.SetFilter("", true)
	co.SetLabelSelector(labels.Everything(), true)
	if f, ok := p.FilterArg(); ok {
		co.SetFilter(f, true)
	}
	if f, ok := p.FuzzyArg(); ok {
		co.SetFilter("-f "+f, true)
	}
	if sel, err := p.LabelsSelector(); err == nil {
		co.SetLabelSelector(sel, false)
	} else {
		slog.Error("Unable to grok labels selector", slogs.Error, err)
	}

	return c.exec(p, gvr, co, clearStack, pushCmd)
}

func (c *Command) defaultCmd(isRoot bool) error {
	if c.app.Conn() == nil || !c.app.Conn().ConnectionOK() {
		return c.run(cmd.NewInterpreter("context"), "", true, true)
	}

	defCmd := podCmd
	if isRoot {
		defCmd = ctxCmd
	}
	p := cmd.NewInterpreter(c.app.Config.ActiveView())
	if p.IsBlank() {
		return c.run(p.Reset(defCmd), "", true, true)
	}

	if err := c.run(p, "", true, true); err != nil {
		slog.Error("Command exec failed. Using default command",
			slogs.Command, p.GetLine(),
			slogs.Error, err,
		)
		p = p.Reset(defCmd)
		return c.run(p, "", true, true)
	}

	return nil
}

func (c *Command) specialCmd(p *cmd.Interpreter, pushCmd bool) bool {
	switch {
	case p.IsCowCmd():
		if msg, ok := p.CowArg(); !ok {
			c.app.Flash().Errf("Invalid command. Use `cow xxx`")
		} else {
			c.app.cowCmd(msg)
		}
	case p.IsBailCmd():
		c.app.BailOut(0)
	case p.IsHelpCmd():
		_ = c.app.helpCmd(nil)
	case p.IsAliasCmd():
		if err := c.aliasCmd(p, pushCmd); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsXrayCmd():
		if err := c.xrayCmd(p, pushCmd); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsRBACCmd():
		if cat, sub, ok := p.RBACArgs(); !ok {
			c.app.Flash().Errf("Invalid command. Use `can [u|g|s]:xxx`")
		} else if err := c.app.inject(NewPolicy(c.app, cat, sub), true); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsContextCmd():
		if err := c.contextCmd(p, pushCmd); err != nil {
			c.app.Flash().Err(err)
		}
	case p.IsNamespaceCmd():
		return c.namespaceCmd(p)
	case p.IsDirCmd():
		if a, ok := p.DirArg(); !ok {
			c.app.Flash().Errf("Invalid command. Use `dir xxx`")
		} else if err := c.app.dirCmd(a, pushCmd); err != nil {
			c.app.Flash().Err(err)
		}
	default:
		return false
	}

	return true
}

func (c *Command) viewMetaFor(p *cmd.Interpreter) (*client.GVR, *MetaViewer, *cmd.Interpreter, error) {
	gvr, ok := c.alias.Resolve(p)
	if !ok {
		return client.NoGVR, nil, nil, fmt.Errorf("`%s` command not found", p.Cmd())
	}

	v := MetaViewer{
		viewerFn: func(gvr *client.GVR) ResourceViewer {
			return NewScaleExtender(NewOwnerExtender(NewBrowser(gvr)))
		},
	}
	if mv, ok := customViewers[gvr]; ok {
		v = mv
	}

	return gvr, &v, p, nil
}

func (*Command) componentFor(gvr *client.GVR, fqn string, v *MetaViewer) ResourceViewer {
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

func (c *Command) exec(p *cmd.Interpreter, gvr *client.GVR, comp model.Component, clearStack, pushCmd bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			slog.Error("Failure detected during command exec", slogs.Error, e)
			c.app.Content.Dump()
			slog.Debug("Dumping history buffer", slogs.CmdHist, c.app.cmdHistory.List())
			slog.Error("Dumping stack", slogs.Stack, string(debug.Stack()))

			ci := cmd.NewInterpreter(podCmd)
			currentCommand, ok := c.app.cmdHistory.Top()
			if ok {
				ci = ci.Reset(currentCommand)
			}
			err = c.run(ci, "", true, true)
		}
	}()

	if comp == nil {
		return fmt.Errorf("no component found for %s", gvr)
	}
	comp.SetCommand(p)

	if clearStack {
		v := contextRX.ReplaceAllString(p.GetLine(), "")
		c.app.Config.SetActiveView(v)
	}
	if err := c.app.inject(comp, clearStack); err != nil {
		return err
	}
	if pushCmd {
		c.app.cmdHistory.Push(p.GetLine())
	}
	slog.Debug("History", slogs.Stack, strings.Join(c.app.cmdHistory.List(), "|"))

	return
}

package views

import (
	"regexp"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

type subjectViewer interface {
	resourceViewer

	setSubject(s string)
}

type command struct {
	app     *appView
	history *ui.CmdStack
}

func newCommand(app *appView) *command {
	return &command{app: app, history: ui.NewCmdStack()}
}

func (c *command) lastCmd() bool {
	return c.history.Last()
}

func (c *command) pushCmd(cmd string) {
	c.history.Push(cmd)
	c.app.Crumbs().Refresh(c.history.Items())
}

func (c *command) previousCmd() (string, bool) {
	c.history.Pop()
	c.app.Crumbs().Refresh(c.history.Items())

	return c.history.Top()
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.pushCmd(c.app.Config.ActiveView())
	c.run(c.app.Config.ActiveView())
}

// Helpers...

var policyMatcher = regexp.MustCompile(`\Apol\s([u|g|s]):([\w-:]+)\b`)

func (c *command) isStdCmd(cmd string) bool {
	switch {
	case cmd == "q", cmd == "quit":
		c.app.BailOut()
		return true
	case cmd == "?", cmd == "help":
		c.app.helpCmd(nil)
		return true
	case cmd == "alias":
		c.app.aliasCmd(nil)
		return true
	case policyMatcher.MatchString(cmd):
		tokens := policyMatcher.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			c.app.inject(newPolicyView(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
	}
	return false
}

func (c *command) isAliasCmd(cmd string) bool {
	cmds := make(map[string]*resourcesCommand, 30)
	resourceViews(c.app.Conn(), cmds)
	res, ok := cmds[cmd]
	if !ok {
		return false
	}

	var r resource.List
	if res.listFn != nil {
		r = res.listFn(c.app.Conn(), resource.DefaultNamespace, res.gvr)
	}

	v := res.viewFn(res.kind, c.app, r)
	if res.colorerFn != nil {
		v.setColorerFn(res.colorerFn)
	}
	if res.enterFn != nil {
		v.setEnterFn(res.enterFn)
	}
	if res.decorateFn != nil {
		v.setDecorateFn(res.decorateFn)
	}

	c.app.Flash().Infof("Viewing resource %s...", res.kind)
	log.Debug().Msgf("Running command %s", cmd)
	c.exec(cmd, v)

	return true
}

func (c *command) isCRDCmd(cmd string) bool {
	crds := map[string]*resourcesCommand{}
	allCRDs(c.app.Conn(), crds)
	res, ok := crds[cmd]
	if !ok {
		return false
	}

	name := res.plural
	if name == "" {
		name = res.singular
	}
	v := newResourceView(
		name,
		c.app,
		resource.NewCustomList(c.app.Conn(), "", res.gvr),
	)
	v.setColorerFn(ui.DefaultColorer)
	c.exec(cmd, v)

	return true
}

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	if c.isStdCmd(cmd) {
		return true
	}

	if c.isAliasCmd(cmd) {
		return true
	}

	if c.isCRDCmd(cmd) {
		return true
	}

	c.app.Flash().Warnf("Huh? `%s` command not found", cmd)
	return false
}

func (c *command) exec(cmd string, v ui.Igniter) {
	if v == nil {
		return
	}

	c.app.Config.SetActiveView(cmd)
	c.app.Config.Save()
	c.app.inject(v)
}

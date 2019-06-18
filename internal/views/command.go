package views

import (
	"regexp"

	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
)

type subjectViewer interface {
	resourceViewer

	setSubject(s string)
}

type command struct {
	app     *appView
	history *cmdStack
}

func newCommand(app *appView) *command {
	return &command{app: app, history: newCmdStack()}
}

func (c *command) lastCmd() bool {
	return c.history.last()
}

func (c *command) pushCmd(cmd string) {
	c.history.push(cmd)
	c.app.crumbs().update(c.history.stack)
}

func (c *command) previousCmd() (string, bool) {
	c.history.pop()
	c.app.crumbs().update(c.history.stack)

	return c.history.top()
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.pushCmd(c.app.config.ActiveView())
	c.run(c.app.config.ActiveView())
}

// Helpers...

var policyMatcher = regexp.MustCompile(`\Apol\s([u|g|s]):([\w-:]+)\b`)

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	var v resourceViewer
	switch {
	case cmd == "q", cmd == "quit":
		c.app.BailOut()
		return true
	case cmd == "?", cmd == "help":
		c.app.inject(newHelpView(c.app, c.app.currentView()))
		return true
	case cmd == "alias":
		c.app.inject(newAliasView(c.app, c.app.currentView()))
		return true
	case policyMatcher.MatchString(cmd):
		tokens := policyMatcher.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			c.app.inject(newPolicyView(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
	default:
		cmds := make(map[string]resCmd, 30)
		resourceViews(c.app.conn(), cmds)
		if res, ok := cmds[cmd]; ok {
			var r resource.List
			if res.listFn != nil {
				r = res.listFn(c.app.conn(), resource.DefaultNamespace)
			}
			v = res.viewFn(res.title, c.app, r)
			if res.colorerFn != nil {
				v.setColorerFn(res.colorerFn)
			}
			if res.enterFn != nil {
				v.setEnterFn(res.enterFn)
			}
			if res.decorateFn != nil {
				v.setDecorateFn(res.decorateFn)
			}
			const fmat = "Viewing resource %s..."
			c.app.flash().infof(fmat, res.title)
			log.Debug().Msgf("Running command %s", cmd)
			c.exec(cmd, v)
			return true
		}
	}

	cmds := make(map[string]resCmd, 30)
	allCRDs(c.app.conn(), cmds)
	res, ok := cmds[cmd]
	if !ok {
		c.app.flash().warnf("Huh? `%s` command not found", cmd)
		return false
	}

	name := res.plural
	if name == "" {
		name = res.singular
	}
	v = newResourceView(
		res.title,
		c.app,
		resource.NewCustomList(c.app.conn(), "", res.api, res.version, name),
	)
	v.setColorerFn(defaultColorer)
	c.exec(cmd, v)

	return true
}

func (c *command) exec(cmd string, v igniter) {
	if v == nil {
		return
	}

	c.app.config.SetActiveView(cmd)
	c.app.config.Save()
	c.app.inject(v)
}

package views

import (
	"fmt"
	"regexp"
	"time"

	"github.com/derailed/k9s/internal/k8s"
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
	cmd := c.app.Config.ActiveView()
	c.pushCmd(cmd)
	if !c.run(cmd) {
		log.Error().Err(fmt.Errorf("Unable to load command %s", cmd)).Msg("Command failed")
	}
}

// Helpers...

var policyMatcher = regexp.MustCompile(`\Apol\s([u|g|s]):([\w-:]+)\b`)

func (c *command) isCustCmd(cmd string) bool {
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

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	defer func(t time.Time) {
		log.Debug().Msgf("RUN CMD Elapsed %v", time.Since(t))
	}(time.Now())

	if c.isCustCmd(cmd) {
		return true
	}

	vv := make(viewers, 200)
	resourceViews(c.app.Conn(), vv)
	allCRDs(c.app.Conn(), vv)
	gvr, ok := aliases.Get(cmd)
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` command not found", cmd)).Msg("Command Failed")
		c.app.Flash().Warnf("Huh? `%s` command not found", cmd)
		return false
	}
	v, ok := vv[gvr]
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` viewer not found", cmd)).Msg("Viewer Failed")
		c.app.Flash().Warnf("Huh? `%s` viewer not found", gvr)
		return false
	}
	return c.execCmd(gvr, v)
}

func (c *command) execCmd(gvr string, v viewer) bool {
	log.Debug().Msgf("ExecCmd gvr %s", gvr)
	var r resource.List
	if v.listFn != nil {
		r = v.listFn(c.app.Conn(), resource.DefaultNamespace)
	}

	var view resourceViewer
	if v.viewFn != nil {
		view = v.viewFn(v.kind, gvr, c.app, r)
	} else {
		view = newResourceView(v.kind, gvr, c.app, r)
	}
	if v.colorerFn != nil {
		view.setColorerFn(v.colorerFn)
	}
	if v.enterFn != nil {
		view.setEnterFn(v.enterFn)
	}
	if v.decorateFn != nil {
		view.setDecorateFn(v.decorateFn)
	}

	return c.exec(gvr, view)
}

func (c *command) exec(gvr string, v ui.Igniter) bool {
	if v == nil {
		log.Error().Err(fmt.Errorf("No igniter given for %s", gvr))
		return false
	}

	g := k8s.GVR(gvr)
	c.app.Flash().Infof("Viewing resource %s...", g.ToR())
	log.Debug().Msgf("Running command %s", gvr)
	c.app.Config.SetActiveView(g.ToR())
	c.app.Config.Save()
	c.app.inject(v)

	return true
}

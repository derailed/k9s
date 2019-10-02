package views

import (
	"fmt"
	"regexp"
	"strings"
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

var authRX = regexp.MustCompile(`\Apol\s([u|g|s]):([\w-:]+)\b`)

func (c *command) isK9sCmd(cmd string) bool {
	cmds := strings.Split(cmd, " ")
	switch cmds[0] {
	case "q", "quit":
		c.app.BailOut()
		return true
	case "?", "help":
		c.app.helpCmd(nil)
		return true
	case "alias":
		c.app.aliasCmd(nil)
		return true
	default:
		if !authRX.MatchString(cmd) {
			return false
		}
		tokens := authRX.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			c.app.inject(newPolicyView(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
	}
	return false
}

// load scrape api for resources and populate aliases.
func (c *command) load() viewers {
	vv := make(viewers, 100)
	resourceViews(c.app.Conn(), vv)
	allCRDs(c.app.Conn(), vv)

	return vv
}

func (c *command) viewMetaFor(cmd string) (string, *viewer) {
	vv := c.load()
	gvr, ok := aliases.Get(cmd)
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` command not found", cmd)).Msg("Command Failed")
		c.app.Flash().Warnf("Huh? `%s` command not found", cmd)
		return "", nil
	}
	v, ok := vv[gvr]
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` viewer not found", gvr)).Msg("Viewer Failed")
		c.app.Flash().Warnf("Huh? viewer for %s not found", cmd)
		return "", nil
	}

	return gvr, &v
}

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	log.Debug().Msgf("Running command %v", cmd)
	defer func(t time.Time) {
		log.Debug().Msgf("RUN CMD Elapsed %v", time.Since(t))
	}(time.Now())

	if c.isK9sCmd(cmd) {
		return true
	}

	cmds := strings.Split(cmd, " ")
	gvr, v := c.viewMetaFor(cmds[0])
	if v == nil {
		return false
	}
	switch cmds[0] {
	case "ctx", "context", "contexts":
		if len(cmds) == 2 {
			c.app.switchCtx(cmds[1], true)
			return true
		}
		view := c.viewerFor(gvr, v)
		return c.exec(gvr, "", view)
	default:
		ns := c.app.Config.ActiveNamespace()
		if len(cmds) == 2 {
			ns = cmds[1]
		}
		if !c.app.switchNS(ns) {
			return false
		}
		return c.exec(gvr, ns, c.viewerFor(gvr, v))
	}

	return false
}

func (c *command) viewerFor(gvr string, v *viewer) resourceViewer {
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

	return view
}

func (c *command) exec(gvr string, ns string, v ui.Igniter) bool {
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

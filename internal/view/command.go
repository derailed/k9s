package view

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
)

var customViewers MetaViewers

type command struct {
	app *App
}

func newCommand(app *App) *command {
	return &command{app: app}
}

func (c *command) Init() error {
	if err := dao.Load(c.app.factory); err != nil {
		return err
	}
	if err := loadAliases(); err != nil {
		return err
	}
	customViewers = loadCustomViewers()

	return nil
}

func (c *command) defaultCmd() {
	cmd := c.app.Config.ActiveView()
	if !c.run(cmd) {
		log.Error().Err(fmt.Errorf("Unable to load command %s", cmd)).Msg("Command failed")
	}
}

var authRX = regexp.MustCompile(`\Apol\s([u|g|s]):([\w-:]+)\b`)

func (c *command) isK9sCmd(cmd string) bool {
	cmds := strings.Split(cmd, " ")
	switch cmds[0] {
	case "q", "Q", "quit":
		c.app.BailOut()
		return true
	case "?", "h", "help":
		c.app.helpCmd(nil)
		return true
	case "a", "alias":
		c.app.aliasCmd(nil)
		return true
	default:
		if !authRX.MatchString(cmd) {
			return false
		}
		tokens := authRX.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			c.app.inject(NewPolicy(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
	}
	return false
}

func (c *command) viewMetaFor(cmd string) (string, *MetaViewer) {
	gvr, ok := aliases.Get(cmd)
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` command not found", cmd)).Msg("Command Failed")
		c.app.Flash().Warnf("Huh? `%s` command not found", cmd)
		return "", nil
	}
	v, ok := customViewers[dao.GVR(gvr)]
	if !ok {
		log.Error().Err(fmt.Errorf("Huh? `%s` viewer not found", gvr)).Msg("MetaViewer Failed")
		c.app.Flash().Warnf("Huh? viewer for %s not found", cmd)
		return "", nil
	}

	return gvr, &v
}

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
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
		if len(cmds) == 2 && c.app.switchCtx(cmds[1], true) != nil {
			log.Error().Msg("Context switch failed!")
			return false
		}
		view := c.componentFor(gvr, v)
		return c.exec(gvr, view)
	default:
		// checks if command includes a namespace
		ns := c.app.Config.ActiveNamespace()
		if len(cmds) == 2 {
			ns = cmds[1]
		}
		if !c.app.switchNS(ns) {
			return false
		}
		return c.exec(gvr, c.componentFor(gvr, v))
	}
}

func (c *command) componentFor(gvr string, v *MetaViewer) ResourceViewer {
	var r resource.List
	if v.listFn != nil {
		r = v.listFn(c.app.Conn(), resource.DefaultNamespace)
	}

	var view ResourceViewer
	if v.viewerFn != nil {
		log.Debug().Msgf("Custom viewer for %s", gvr)
		view = v.viewerFn(dao.GVR(gvr))
		// view = NewGeneric(dao.GVR(gvr))
	} else {
		log.Debug().Msgf("Standard viewer for %s", gvr)
		view = NewResource("BLAH", gvr, r)
	}

	switch o := view.(type) {
	case TableViewer:
		if v.enterFn != nil {
			o.GetTable().SetEnterFn(v.enterFn)
		}
	}

	return view
}

func (c *command) exec(gvr string, comp model.Component) bool {
	if comp == nil {
		log.Error().Err(fmt.Errorf("No component given for %s", gvr))
		return false
	}

	g := k8s.GVR(gvr)
	c.app.Flash().Infof("Viewing %s resource...", g.ToR())
	log.Debug().Msgf("Running command %s", gvr)
	c.app.Config.SetActiveView(g.ToR())
	if err := c.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	c.app.Content.Stack.ClearHistory()
	c.app.inject(comp)

	return true
}

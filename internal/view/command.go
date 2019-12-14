package view

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
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

func (c *command) defaultCmd() error {
	return c.run(c.app.Config.ActiveView())
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
			// BOZO!!
			// c.app.inject(NewPolicy(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
	}
	return false
}

func (c *command) viewMetaFor(cmd string) (string, *MetaViewer, error) {
	gvr, ok := aliases.Get(cmd)
	if !ok {
		return "", nil, fmt.Errorf("Huh? `%s` command not found", cmd)
	}
	v, ok := customViewers[client.GVR(gvr)]
	if !ok {
		return gvr, &MetaViewer{viewerFn: NewBrowser}, nil
	}

	return gvr, &v, nil
}

// Exec the command by showing associated display.
func (c *command) run(cmd string) error {
	if c.isK9sCmd(cmd) {
		return nil
	}

	cmds := strings.Split(cmd, " ")
	gvr, v, err := c.viewMetaFor(cmds[0])
	if err != nil {
		return err
	}
	switch cmds[0] {
	case "ctx", "context", "contexts":
		if len(cmds) == 2 && c.app.switchCtx(cmds[1], true) != nil {
			return fmt.Errorf("context switch failed!")
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
			return fmt.Errorf("namespace switch failed for ns %q", ns)
		}
		return c.exec(gvr, c.componentFor(gvr, v))
	}
}

func (c *command) componentFor(gvr string, v *MetaViewer) ResourceViewer {
	var view ResourceViewer
	if v.viewerFn != nil {
		log.Debug().Msgf("Custom viewer for %s", gvr)
		view = v.viewerFn(client.GVR(gvr))
	} else {
		log.Debug().Msgf("Generic viewer for %s", gvr)
		view = NewBrowser(client.GVR(gvr))
	}

	if v.enterFn != nil {
		log.Debug().Msgf("SETTING CUSTOM ENTER ON %s", gvr)
		view.GetTable().SetEnterFn(v.enterFn)
	}

	return view
}

func (c *command) exec(gvr string, comp model.Component) error {
	if comp == nil {
		return fmt.Errorf("No component given for %s", gvr)
	}

	g := client.GVR(gvr)
	c.app.Flash().Infof("Viewing %s resource...", g.ToR())
	log.Debug().Msgf("Running command %s", gvr)
	c.app.Config.SetActiveView(g.ToR())
	if err := c.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	c.app.Content.Stack.ClearHistory()
	return c.app.inject(comp)

	return nil
}

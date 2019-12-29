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

type Command struct {
	app *App

	alias *dao.Alias
}

func NewCommand(app *App) *Command {
	return &Command{
		app: app,
	}
}

func (c *Command) Init() error {
	c.alias = dao.NewAlias(c.app.factory)
	if _, err := c.alias.Ensure(); err != nil {
		return err
	}
	customViewers = loadCustomViewers()

	return nil
}

// Reset resets Command and reload aliases.
func (c *Command) Reset() error {
	c.alias.Clear()
	if _, err := c.alias.Ensure(); err != nil {
		return err
	}

	return nil
}

func (c *Command) defaultCmd() error {
	return c.run(c.app.Config.ActiveView())
}

var canRX = regexp.MustCompile(`\Acan\s([u|g|s]):([\w-:]+)\b`)

func (c *Command) specialCmd(cmd string) bool {
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
		if !canRX.MatchString(cmd) {
			return false
		}
		tokens := canRX.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			if err := c.app.inject(NewPolicy(c.app, tokens[0][1], tokens[0][2])); err != nil {
				log.Error().Err(err).Msgf("policy view load failed")
				return false
			}
			return true
		}
	}
	return false
}

func (c *Command) viewMetaFor(cmd string) (string, *MetaViewer, error) {
	gvr, ok := c.alias.Get(cmd)
	if !ok {
		return "", nil, fmt.Errorf("Huh? `%s` Command not found", cmd)
	}

	v, ok := customViewers[client.GVR(gvr)]
	if !ok {
		return gvr, &MetaViewer{viewerFn: NewBrowser}, nil
	}

	return gvr, &v, nil
}

// Exec the Command by showing associated display.
func (c *Command) run(cmd string) error {
	if c.specialCmd(cmd) {
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
		// checks if Command includes a namespace
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

func (c *Command) componentFor(gvr string, v *MetaViewer) ResourceViewer {
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

func (c *Command) exec(gvr string, comp model.Component) error {
	if comp == nil {
		return fmt.Errorf("No component given for %s", gvr)
	}

	g := client.GVR(gvr)
	c.app.Flash().Infof("Viewing %s resource...", g.ToR())
	log.Debug().Msgf("Running Command %s", gvr)
	c.app.Config.SetActiveView(g.ToR())
	if err := c.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	c.app.Content.Stack.ClearHistory()

	return c.app.inject(comp)
}

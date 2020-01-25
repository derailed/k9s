package view

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	customViewers MetaViewers

	canRX = regexp.MustCompile(`\Acan\s([u|g|s]):([\w-:]+)\b`)
)

// Command represents a user command.
type Command struct {
	app *App

	alias *dao.Alias
	mx    sync.Mutex
}

// NewCommand returns a new command.
func NewCommand(app *App) *Command {
	return &Command{
		app: app,
	}
}

// Init initializes the command.
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
	c.mx.Lock()
	defer c.mx.Unlock()

	c.alias.Clear()
	if _, err := c.alias.Ensure(); err != nil {
		return err
	}

	return nil
}

func allowedXRay(gvr client.GVR) bool {
	gg := []string{
		"v1/pods",
		"v1/services",
		"apps/v1/deployments",
		"apps/v1/daemonsets",
		"apps/v1/statefulsets",
		"apps/v1/replicasets",
	}
	for _, g := range gg {
		if g == gvr.String() {
			return true
		}
	}

	return false
}

func (c *Command) xrayCmd(cmd string) error {
	tokens := strings.Split(cmd, " ")
	if len(tokens) < 2 {
		return errors.New("You must specify a resource")
	}
	gvr, ok := c.alias.AsGVR(tokens[1])
	if !ok {
		return fmt.Errorf("Huh? `%s` Command not found", cmd)
	}
	if !allowedXRay(gvr) {
		return fmt.Errorf("Huh? `%s` Command not found", cmd)
	}

	x := NewXray(gvr)
	ns := c.app.Config.ActiveNamespace()
	if len(tokens) == 3 {
		ns = tokens[2]
	}
	if err := c.app.Config.SetActiveNamespace(client.CleanseNamespace(ns)); err != nil {
		return err
	}
	if err := c.app.Config.Save(); err != nil {
		return err
	}

	return c.exec(cmd, "xrays", x, true)
}

// Exec the Command by showing associated display.
func (c *Command) run(cmd, path string, clearStack bool) error {
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
		if len(cmds) == 2 {
			return useContext(c.app, cmds[1])
		}
		view := c.componentFor(gvr, path, v)
		return c.exec(cmd, gvr, view, clearStack)
	default:
		// checks if Command includes a namespace
		ns := c.app.Config.ActiveNamespace()
		if len(cmds) == 2 {
			ns = cmds[1]
		}
		if !c.app.switchNS(ns) {
			return fmt.Errorf("namespace switch failed for ns %q", ns)
		}
		if !c.alias.Check(cmds[0]) {
			return fmt.Errorf("Huh? `%s` Command not found", cmd)
		}
		return c.exec(cmd, gvr, c.componentFor(gvr, path, v), clearStack)
	}
}

func (c *Command) defaultCmd() error {
	err := c.run(c.app.Config.ActiveView(), "", true)
	if err != nil {
		log.Error().Err(err).Msgf("Saved command failed. Loading default view")
		return c.run("pod", "", true)
	}
	return nil
}

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
	case "x", "xray":
		if err := c.xrayCmd(cmd); err != nil {
			c.app.Flash().Err(err)
		}
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
	gvr, ok := c.alias.AsGVR(cmd)
	if !ok {
		return "", nil, fmt.Errorf("Huh? `%s` Command not found", cmd)
	}

	v, ok := customViewers[gvr]
	if !ok {
		return gvr.String(), &MetaViewer{viewerFn: NewBrowser}, nil
	}

	return gvr.String(), &v, nil
}

func (c *Command) componentFor(gvr, path string, v *MetaViewer) ResourceViewer {
	var view ResourceViewer
	if v.viewerFn != nil {
		view = v.viewerFn(client.NewGVR(gvr))
	} else {
		view = NewBrowser(client.NewGVR(gvr))
	}

	view.SetInstance(path)
	if v.enterFn != nil {
		view.GetTable().SetEnterFn(v.enterFn)
	}

	return view
}

func (c *Command) exec(cmd, gvr string, comp model.Component, clearStack bool) error {
	if comp == nil {
		return fmt.Errorf("No component found for %s", gvr)
	}
	c.app.Flash().Infof("Viewing %s...", client.NewGVR(gvr).R())
	c.app.Config.SetActiveView(cmd)
	if err := c.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	if clearStack {
		c.app.Content.Stack.Clear()
	}

	return c.app.inject(comp)
}

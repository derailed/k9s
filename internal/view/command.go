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
		log.Error().Err(err).Msgf("command init failed!")
		return err
	}
	customViewers = loadCustomViewers()

	return nil
}

// Reset resets Command and reload aliases.
func (c *Command) Reset(clear bool) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if clear {
		c.alias.Clear()
	}
	if _, err := c.alias.Ensure(); err != nil {
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

func (c *Command) xrayCmd(cmd string) error {
	tokens := strings.Split(cmd, " ")
	if len(tokens) < 2 {
		return errors.New("you must specify a resource")
	}
	gvr, ok := c.alias.AsGVR(tokens[1])
	if !ok {
		return fmt.Errorf("`%s` command not found", cmd)
	}
	if !allowedXRay(gvr) {
		return fmt.Errorf("`%s` command not found", cmd)
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

// Run execs the command by showing associated display.
func (c *Command) run(cmd, path string, clearStack bool) error {
	if c.specialCmd(cmd, path) {
		return nil
	}
	cmds := strings.Split(cmd, " ")
	command := strings.ToLower(cmds[0])
	gvr, v, err := c.viewMetaFor(command)
	if err != nil {
		return err
	}
	var cns string
	tt := strings.Split(gvr, " ")
	if len(tt) == 2 {
		gvr, cns = tt[0], tt[1]
	}

	switch command {
	case "ctx", "context", "contexts":
		if len(cmds) == 2 {
			return useContext(c.app, cmds[1])
		}
		return c.exec(cmd, gvr, c.componentFor(gvr, path, v), clearStack)
	case "dir":
		if len(cmds) != 2 {
			return errors.New("you must specify a directory")
		}
		return c.app.dirCmd(cmds[1])
	default:
		// checks if Command includes a namespace
		ns := c.app.Config.ActiveNamespace()
		if len(cmds) == 2 {
			ns = cmds[1]
		}
		if cns != "" {
			ns = cns
		}
		if err := c.app.switchNS(ns); err != nil {
			return err
		}
		if !c.alias.Check(command) {
			return fmt.Errorf("`%s` Command not found", cmd)
		}
		return c.exec(cmd, gvr, c.componentFor(gvr, path, v), clearStack)
	}
}

func (c *Command) defaultCmd() error {
	if c.app.Conn() == nil || !c.app.Conn().ConnectionOK() {
		return c.run("context", "", true)
	}
	view := c.app.Config.ActiveView()
	if view == "" {
		return c.run("pod", "", true)
	}
	tokens := strings.Split(view, " ")
	cmd := view
	if len(tokens) == 1 {
		if !isContextCmd(tokens[0]) {
			cmd = tokens[0] + " " + c.app.Config.ActiveNamespace()
		}
	}

	if err := c.run(cmd, "", true); err != nil {
		log.Error().Err(err).Msgf("Default run command failed %q", cmd)
		return c.run("pod", "", true)
	}
	return nil
}

func isContextCmd(c string) bool {
	return c == "ctx" || c == "context"
}

func (c *Command) specialCmd(cmd, path string) bool {
	cmds := strings.Split(cmd, " ")
	switch cmds[0] {
	case "cow":
		c.app.cowCmd(path)
		return true
	case "q", "q!", "qa", "Q", "quit":
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
			if err := c.app.inject(NewPolicy(c.app, tokens[0][1], tokens[0][2]), false); err != nil {
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
		return "", nil, fmt.Errorf("`%s` command not found", cmd)
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

func (c *Command) exec(cmd, gvr string, comp model.Component, clearStack bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error().Msgf("Something bad happened! %#v", e)
			c.app.Content.Dump()
			log.Debug().Msgf("History %v", c.app.cmdHistory.List())
			log.Error().Msg(string(debug.Stack()))

			hh := c.app.cmdHistory.List()
			if len(hh) == 0 {
				_ = c.run("pod", "", true)
			} else {
				_ = c.run(hh[0], "", true)
			}
			err = fmt.Errorf("invalid command %q", cmd)
		}
	}()

	if comp == nil {
		return fmt.Errorf("no component found for %s", gvr)
	}
	c.app.Flash().Infof("Viewing %s...", client.NewGVR(gvr).R())
	command := cmd
	if tokens := strings.Split(cmd, " "); len(tokens) >= 2 {
		command = tokens[0]
	}
	c.app.Config.SetActiveView(command)
	if err := c.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	if err := c.app.inject(comp, clearStack); err != nil {
		return err
	}

	c.app.cmdHistory.Push(cmd)

	return
}

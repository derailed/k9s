package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Runner represents a runnable action handler.
type Runner interface {
	App() *App
	GetSelectedItem() string
	Aliases() []string
	EnvFn() EnvFunc
}

func hasAll(scopes []string) bool {
	for _, s := range scopes {
		if s == "all" {
			return true
		}
	}
	return false
}

func includes(aliases []string, s string) bool {
	for _, a := range aliases {
		if a == s {
			return true
		}
	}
	return false
}

func inScope(scopes, aliases []string) bool {
	if hasAll(scopes) {
		return true
	}
	for _, s := range scopes {
		if includes(aliases, s) {
			return true
		}
	}

	return false
}

func hotKeyActions(r Runner, aa ui.KeyActions) {
	hh := config.NewHotKeys()
	if err := hh.Load(); err != nil {
		return
	}

	for k, hk := range hh.HotKey {
		key, err := asKey(hk.ShortCut)
		if err != nil {
			log.Warn().Err(err).Msg("HOT-KEY Unable to map hotkey shortcut to a key")
			continue
		}
		_, ok := aa[key]
		if ok {
			log.Warn().Err(fmt.Errorf("HOT-KEY Doh! you are trying to overide an existing command `%s", k)).Msg("Invalid shortcut")
			continue
		}
		aa[key] = ui.NewSharedKeyAction(
			hk.Description,
			gotoCmd(r, hk.Command),
			false)
	}
}

func gotoCmd(r Runner, cmd string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if err := r.App().gotoResource(cmd, true); err != nil {
			r.App().Flash().Err(err)
		}
		return nil
	}
}

func pluginActions(r Runner, aa ui.KeyActions) {
	pp := config.NewPlugins()
	if err := pp.Load(); err != nil {
		return
	}

	for k, plugin := range pp.Plugin {
		if !inScope(plugin.Scopes, r.Aliases()) {
			continue
		}
		key, err := asKey(plugin.ShortCut)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to map plugin shortcut to a key")
			continue
		}
		_, ok := aa[key]
		if ok {
			log.Warn().Err(fmt.Errorf("Doh! you are trying to overide an existing command `%s", k)).Msg("Invalid shortcut")
			continue
		}
		aa[key] = ui.NewKeyAction(
			plugin.Description,
			execCmd(r, plugin.Command, plugin.Background, plugin.Args...),
			true)
	}
}

func execCmd(r Runner, bin string, bg bool, args ...string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := r.GetSelectedItem()
		if path == "" {
			return evt
		}

		ns, _ := client.Namespaced(path)
		var (
			env = r.EnvFn()()
			aa  = make([]string, len(args))
			err error
		)
		for i, a := range args {
			aa[i], err = env.envFor(ns, a)
			if err != nil {
				log.Error().Err(err).Msg("Plugin Args match failed")
				return nil
			}
		}
		if run(true, r.App(), bin, bg, aa...) {
			r.App().Flash().Info("Plugin command launched successfully!")
		} else {
			r.App().Flash().Info("Plugin command failed!")
		}

		return nil
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// AllScopes represents actions available for all views.
const AllScopes = "all"

// Runner represents a runnable action handler.
type Runner interface {
	App() *App
	GetSelectedItem() string
	Aliases() map[string]struct{}
	EnvFn() EnvFunc
}

func hasAll(scopes []string) bool {
	for _, s := range scopes {
		if s == AllScopes {
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

func inScope(scopes []string, aliases map[string]struct{}) bool {
	if hasAll(scopes) {
		return true
	}
	for _, s := range scopes {
		if _, ok := aliases[s]; ok {
			return ok
		}
	}

	return false
}

func hotKeyActions(r Runner, aa ui.KeyActions) error {
	hh := config.NewHotKeys()
	for k, a := range aa {
		if a.Opts.HotKey {
			delete(aa, k)
		}
	}

	var errs error
	if err := hh.Load(r.App().Config.ContextHotkeysPath()); err != nil {
		errs = errors.Join(errs, err)
	}
	for k, hk := range hh.HotKey {
		key, err := asKey(hk.ShortCut)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		_, ok := aa[key]
		if ok && !hk.Override {
			errs = errors.Join(errs, fmt.Errorf("duplicated hotkeys found for %q in %q", hk.ShortCut, k))
			continue
		} else if ok && hk.Override == true {
			log.Info().Msgf("Action %q has been overrided by hotkey in %q", hk.ShortCut, k)
		}

		command, err := r.EnvFn()().Substitute(hk.Command)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid shortcut command")
			continue
		}

		aa[key] = ui.NewKeyActionWithOpts(
			hk.Description,
			gotoCmd(r, command, "", !hk.KeepHistory),
			ui.ActionOpts{
				Shared: true,
				HotKey: true,
			},
		)
	}

	return errs
}

func gotoCmd(r Runner, cmd, path string, clearStack bool) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		r.App().gotoResource(cmd, path, clearStack)
		return nil
	}
}

func pluginActions(r Runner, aa ui.KeyActions) error {
	pp := config.NewPlugins()
	for k, a := range aa {
		if a.Opts.Plugin {
			delete(aa, k)
		}
	}

	var errs error
	if err := pp.Load(r.App().Config.ContextPluginsPath()); err != nil {
		errs = errors.Join(errs, err)
	}
	aliases := r.Aliases()
	for k, plugin := range pp.Plugins {
		if !inScope(plugin.Scopes, aliases) {
			continue
		}
		key, err := asKey(plugin.ShortCut)

		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		_, ok := aa[key]
		if ok && !plugin.Override {
			errs = errors.Join(errs, fmt.Errorf("duplicated plugin key found for %q in %q", plugin.ShortCut, k))
			continue
		} else if ok && plugin.Override == true {
			log.Info().Msgf("Action %q has been overrided by plugin in %q", plugin.ShortCut, k)
		}
		aa[key] = ui.NewKeyActionWithOpts(
			plugin.Description,
			pluginAction(r, plugin),
			ui.ActionOpts{
				Visible: true,
				Plugin:  true,
			})
	}

	return errs
}

func pluginAction(r Runner, p config.Plugin) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := r.GetSelectedItem()
		if path == "" {
			return evt
		}
		if r.EnvFn() == nil {
			return nil
		}

		args := make([]string, len(p.Args))
		for i, a := range p.Args {
			arg, err := r.EnvFn()().Substitute(a)
			if err != nil {
				log.Error().Err(err).Msg("Plugin Args match failed")
				return nil
			}
			args[i] = arg
		}

		cb := func() {
			opts := shellOpts{
				binary:     p.Command,
				background: p.Background,
				pipes:      p.Pipes,
				args:       args,
			}
			suspend, errChan, statusChan := run(r.App(), opts)
			if !suspend {
				r.App().Flash().Infof("Plugin command failed: %q", p.Description)
				return
			}
			var errs error
			for e := range errChan {
				errs = errors.Join(errs, e)
			}
			if errs != nil {
				r.App().cowCmd(errs.Error())
				return
			}
			go func() {
				for st := range statusChan {
					r.App().Flash().Infof("Plugin command launched successfully: %q", st)
				}
			}()

		}
		if p.Confirm {
			msg := fmt.Sprintf("Run?\n%s %s", p.Command, strings.Join(args, " "))
			dialog.ShowConfirm(r.App().Styles.Dialog(), r.App().Content.Pages, "Confirm "+p.Description, msg, cb, func() {})
			return nil
		}
		cb()

		return nil
	}
}

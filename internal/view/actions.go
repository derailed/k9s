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

func hotKeyActions(r Runner, aa *ui.KeyActions) error {
	hh := config.NewHotKeys()
	aa.Range(func(k tcell.Key, a ui.KeyAction) {
		if a.Opts.HotKey {
			aa.Delete(k)
		}
	})

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
		if _, ok := aa.Get(key); ok {
			if !hk.Override {
				errs = errors.Join(errs, fmt.Errorf("duplicate hotkey found for %q in %q", hk.ShortCut, k))
				continue
			}
			log.Debug().Msgf("Action %q has been overridden by hotkey in %q", hk.ShortCut, k)
		}

		command, err := r.EnvFn()().Substitute(hk.Command)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid shortcut command")
			continue
		}

		aa.Add(key, ui.NewKeyActionWithOpts(
			hk.Description,
			gotoCmd(r, command, "", !hk.KeepHistory),
			ui.ActionOpts{
				Shared: true,
				HotKey: true,
			},
		))
	}

	return errs
}

func gotoCmd(r Runner, cmd, path string, clearStack bool) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		r.App().gotoResource(cmd, path, clearStack)
		return nil
	}
}

func pluginActions(r Runner, aa *ui.KeyActions) error {
	aa.Range(func(k tcell.Key, a ui.KeyAction) {
		if a.Opts.Plugin {
			aa.Delete(k)
		}
	})

	path, err := r.App().Config.ContextPluginsPath()
	if err != nil {
		return err
	}
	pp := config.NewPlugins()
	if err := pp.Load(path); err != nil {
		return err
	}

	var (
		errs    error
		aliases = r.Aliases()
		ro      = r.App().Config.K9s.IsReadOnly()
	)
	for k, plugin := range pp.Plugins {
		if !inScope(plugin.Scopes, aliases) {
			continue
		}
		key, err := asKey(plugin.ShortCut)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if _, ok := aa.Get(key); ok {
			if !plugin.Override {
				errs = errors.Join(errs, fmt.Errorf("duplicate plugin key found for %q in %q", plugin.ShortCut, k))
				continue
			}
			log.Debug().Msgf("Action %q has been overridden by plugin in %q", plugin.ShortCut, k)
		}

		if plugin.Dangerous && ro {
			continue
		}
		aa.Add(key, ui.NewKeyActionWithOpts(
			plugin.Description,
			pluginAction(r, plugin),
			ui.ActionOpts{
				Visible:   true,
				Plugin:    true,
				Dangerous: plugin.Dangerous,
			},
		))
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
					if !p.OverwriteOutput {
						r.App().Flash().Infof("Plugin command launched successfully: %q", st)
					} else if strings.Contains(st, outputPrefix) {
						infoMsg := strings.TrimPrefix(st, outputPrefix)
						r.App().Flash().Info(strings.TrimSpace(infoMsg))
						return
					}
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

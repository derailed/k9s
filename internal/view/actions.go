// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/util/sets"
)

// AllScopes represents actions available for all views.
const AllScopes = "all"

// Runner represents a runnable action handler.
type Runner interface {
	// App returns the current app.
	App() *App

	// GetSelectedItem returns the current selected item.
	GetSelectedItem() string

	// Aliases returns all aliases assoxciated with the view GVR.
	Aliases() sets.Set[string]

	// EnvFn returns the current environment function.
	EnvFn() EnvFunc
}

func hasAll(scopes []string) bool {
	return slices.Contains(scopes, AllScopes)
}

func includes(aliases []string, s string) bool {
	return slices.Contains(aliases, s)
}

func inScope(scopes []string, aliases sets.Set[string]) bool {
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
			slog.Debug("HotKey overrode action shortcut",
				slogs.Shortcut, hk.ShortCut,
				slogs.Key, k,
			)
		}

		command, err := r.EnvFn()().Substitute(hk.Command)
		if err != nil {
			slog.Warn("Invalid shortcut command", slogs.Error, err)
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
	return func(*tcell.EventKey) *tcell.EventKey {
		r.App().gotoResource(cmd, path, clearStack, true)
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
	if err := pp.Load(path, true); err != nil {
		return err
	}

	var (
		errs    error
		aliases = r.Aliases()
		ro      = r.App().Config.IsReadOnly()
	)
	for k := range pp.Plugins {
		if !inScope(pp.Plugins[k].Scopes, aliases) || (ro && pp.Plugins[k].Dangerous) {
			continue
		}
		key, err := asKey(pp.Plugins[k].ShortCut)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if _, ok := aa.Get(key); ok {
			if !pp.Plugins[k].Override {
				errs = errors.Join(errs, fmt.Errorf("duplicate plugin key found for %q in %q", pp.Plugins[k].ShortCut, k))
				continue
			}
			slog.Debug("Plugin overrode action shortcut",
				slogs.Plugin, k,
				slogs.Key, pp.Plugins[k].ShortCut,
			)
		}

		plugin := pp.Plugins[k]
		aa.Add(key, ui.NewKeyActionWithOpts(
			pp.Plugins[k].Description,
			pluginAction(r, &plugin),
			ui.ActionOpts{
				Visible:   true,
				Plugin:    true,
				Dangerous: plugin.Dangerous,
			},
		))
	}

	return errs
}

func pluginAction(r Runner, p *config.Plugin) ui.ActionHandler {
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
				slog.Error("Plugin Args match failed", slogs.Error, err)
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
			suspend, errChan, statusChan := run(r.App(), &opts)
			if !suspend {
				r.App().Flash().Infof("Plugin command failed: %q", p.Description)
				return
			}
			var errs error
			for e := range errChan {
				errs = errors.Join(errs, e)
			}
			if errs != nil {
				if !strings.Contains(errs.Error(), "signal: interrupt") {
					slog.Error("Plugin command failed", slogs.Error, errs)
					r.App().cowCmd(errs.Error())
					return
				}
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
			d := r.App().Styles.Dialog()
			dialog.ShowConfirm(&d, r.App().Content.Pages, "Confirm "+p.Description, msg, cb, func() {})
			return nil
		}
		cb()

		return nil
	}
}

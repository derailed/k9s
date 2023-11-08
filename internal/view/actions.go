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
	Aliases() []string
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
			log.Warn().Err(fmt.Errorf("HOT-KEY Doh! you are trying to override an existing command `%s", k)).Msg("Invalid shortcut")
			continue
		}
		aa[key] = ui.NewSharedKeyAction(
			hk.Description,
			gotoCmd(r, hk.Command, ""),
			false)
	}
}

func gotoCmd(r Runner, cmd, path string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		r.App().gotoResource(cmd, path, true)
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
			log.Warn().Msgf("Invalid shortcut. You are trying to override an existing command `%s", k)
			continue
		}
		aa[key] = ui.NewKeyAction(
			plugin.Description,
			pluginAction(r, plugin),
			true)
	}
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
				clear:      true,
				binary:     p.Command,
				background: p.Background,
				pipes:      p.Pipes,
				args:       args,
			}
			suspend, errChan := run(r.App(), opts)
			if !suspend {
				r.App().Flash().Info("Plugin command failed!")
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
			r.App().Flash().Info("Plugin command launched successfully!")
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

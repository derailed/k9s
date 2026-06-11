// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

type skinScope int

const (
	skinScopeGlobal skinScope = iota
	skinScopeContext
	defaultSkinOption = "Default / Stock"
)

type skinApplyResult struct {
	scope    skinScope
	skin     string
	warnings []string
}

func (r skinApplyResult) notice() string {
	var b strings.Builder

	switch {
	case r.skin == "" && r.scope == skinScopeGlobal:
		b.WriteString("Default skin restored in the global config and applied.")
	case r.skin == "" && r.scope == skinScopeContext:
		b.WriteString("Default skin restored for the current context and applied.")
	case r.scope == skinScopeGlobal:
		fmt.Fprintf(&b, "Skin %q saved to the global config and applied.", r.skin)
	default:
		fmt.Fprintf(&b, "Skin %q saved to the current context and applied.", r.skin)
	}
	b.WriteString("\n\nRestart K9s for a clean full-session refresh.")
	if len(r.warnings) > 0 {
		b.WriteString("\n\nNotes:")
		for _, warning := range r.warnings {
			b.WriteString("\n- ")
			b.WriteString(warning)
		}
	}

	return b.String()
}

func (a *App) skinPickerCmd(*tcell.EventKey) *tcell.EventKey {
	a.showSkinScopePicker()
	return nil
}

func (a *App) showSkinScopePicker() {
	styles := a.Styles.Dialog()
	options := []string{
		"Global (all contexts)",
		"Current context only",
	}
	dialog.ShowSelection(&styles, a.Content.Pages, "Skin Scope", options, func(index int) {
		switch index {
		case 0:
			a.showSkinPicker(skinScopeGlobal)
		case 1:
			a.showSkinPicker(skinScopeContext)
		}
	})
}

func (a *App) showSkinPicker(scope skinScope) {
	skins, err := installedSkinNames(config.AppSkinsDir)
	if err != nil {
		a.Flash().Err(err)
		return
	}
	if len(skins) == 0 {
		a.Flash().Warnf("No installed skins found in %q", config.AppSkinsDir)
		return
	}

	options := append([]string{defaultSkinOption}, skins...)
	title := "Global Skin"
	if scope == skinScopeContext {
		title = "Context Skin"
	}
	styles := a.Styles.Dialog()
	dialog.ShowSelection(&styles, a.Content.Pages, title, options, func(index int) {
		if index < 0 || index >= len(options) {
			return
		}

		skin := ""
		if index > 0 {
			skin = skins[index-1]
		}
		result, err := a.applySkinSelection(scope, skin)
		if err != nil {
			a.Flash().Err(err)
			return
		}
		dialog.ShowNotice(&styles, a.Content.Pages, "Restart Recommended", result.notice())
	})
}

func installedSkinNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	skins := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		skins = append(skins, strings.TrimSuffix(entry.Name(), ".yaml"))
	}
	sort.Strings(skins)

	return skins, nil
}

func (a *App) applySkinSelection(scope skinScope, skin string) (skinApplyResult, error) {
	result := skinApplyResult{
		scope: scope,
		skin:  skin,
	}
	if envSkin := os.Getenv("K9S_SKIN"); envSkin != "" {
		result.warnings = append(result.warnings,
			fmt.Sprintf("K9S_SKIN=%q still takes precedence over saved config values.", envSkin),
		)
	}

	switch scope {
	case skinScopeGlobal:
		if err := a.saveGlobalSkin(skin); err != nil {
			return skinApplyResult{}, err
		}
		if ct, err := a.Config.CurrentContext(); err == nil && ct.Skin != "" {
			result.warnings = append(result.warnings,
				"The current context has a context-specific skin and still overrides the global selection.",
			)
		}
	case skinScopeContext:
		if err := a.saveContextSkin(skin); err != nil {
			return skinApplyResult{}, err
		}
	default:
		return skinApplyResult{}, fmt.Errorf("unsupported skin scope: %d", scope)
	}

	a.ReloadStyles()

	return result, nil
}

func (a *App) saveGlobalSkin(skin string) error {
	a.Config.K9s.UI.Skin = skin
	return a.Config.SaveFile(config.AppConfigFile)
}

func (a *App) saveContextSkin(skin string) error {
	ct, err := a.Config.CurrentContext()
	if err != nil {
		return err
	}
	ct.Skin = skin
	return a.Config.Save(true)
}

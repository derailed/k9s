// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

const (
	pluginAddGlobalLabel  = "Global (all contexts)"
	pluginAddContextLabel = "Current context only"
	pluginSeedYAML        = `plugins:
  plugin-name:
    shortCut: Shift-X
    description: Describe your plugin
    scopes:
      - all
    command: kubectl
    args:
      - version
`
)

// Plugin presents the effective plugin catalog.
type Plugin struct {
	ResourceViewer
}

// NewPlugin returns a new plugin catalog viewer.
func NewPlugin(gvr *client.GVR) ResourceViewer {
	p := Plugin{
		ResourceViewer: NewBrowser(gvr),
	}
	p.GetTable().SetBorderFocusColor(tcell.ColorMediumAquamarine)
	p.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorMediumAquamarine).Attributes(tcell.AttrNone))
	p.AddBindKeysFn(p.bindKeys)
	p.SetContextFn(p.pluginContext)

	return &p
}

// Init initializes the view.
func (p *Plugin) Init(ctx context.Context) error {
	if err := p.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	p.GetTable().GetModel().SetNamespace(client.NotNamespaced)

	return nil
}

func (p *Plugin) pluginContext(ctx context.Context) context.Context {
	path, err := p.App().Config.ContextPluginsPath()
	if err == nil && path != "" {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
	}

	return ctx
}

func (p *Plugin) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL)
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("View", p.viewCmd, true),
		ui.KeyY:        ui.NewKeyAction(yamlAction, p.viewCmd, true),
		ui.KeyShiftN:   ui.NewKeyAction("Sort Name", p.GetTable().SortColCmd("NAME", true), false),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Shortcut", p.GetTable().SortColCmd("SHORTCUT", true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Command", p.GetTable().SortColCmd("COMMAND", true), false),
		ui.KeyShiftS:   ui.NewKeyAction("Sort Source", p.GetTable().SortColCmd("SOURCE", true), false),
	})
	if !p.App().Config.IsReadOnly() {
		aa.Add(ui.KeyA, ui.NewKeyAction("Add", p.addCmd, true))
		aa.Add(ui.KeyE, ui.NewKeyAction("Edit", p.editCmd, true))
	}
}

func (p *Plugin) selectedPlugin() (path, name string, ok bool) {
	id := p.GetTable().GetSelectedItem()
	if id == "" {
		return "", "", false
	}

	path, name = render.ParsePluginRowID(id)
	return path, name, path != "" && name != ""
}

func (p *Plugin) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	path, name, ok := p.selectedPlugin()
	if !ok {
		return evt
	}

	bb, err := os.ReadFile(path)
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}

	details := NewDetails(p.App(), yamlAction, name, contentYAML, true).Update(string(bb))
	if err := p.App().inject(details, false); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *Plugin) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path, _, ok := p.selectedPlugin()
	if !ok {
		return evt
	}

	p.editFile(path)

	return nil
}

func (p *Plugin) addCmd(evt *tcell.EventKey) *tcell.EventKey {
	type target struct {
		label string
		path  string
	}

	targets := []target{{
		label: pluginAddGlobalLabel,
		path:  config.AppPluginsFile,
	}}
	if path, err := p.App().Config.ContextPluginsPath(); err == nil && path != "" {
		targets = append(targets, target{
			label: pluginAddContextLabel,
			path:  path,
		})
	}

	openTarget := func(t target) {
		if err := seedPluginFile(t.path); err != nil {
			p.App().Flash().Err(err)
			return
		}
		p.editFile(t.path)
	}

	if len(targets) == 1 {
		openTarget(targets[0])
		return nil
	}

	options := make([]string, 0, len(targets))
	for _, t := range targets {
		options = append(options, t.label)
	}
	d := p.App().Styles.Dialog()
	dialog.ShowSelection(&d, p.App().Content.Pages, "Add Plugin", options, func(index int) {
		if index < 0 || index >= len(targets) {
			return
		}
		openTarget(targets[index])
	})

	return nil
}

func (p *Plugin) editFile(path string) {
	p.Stop()
	defer p.Start()
	if !edit(p.App(), &shellOpts{clear: true, args: []string{path}}) {
		p.App().Flash().Errf("Failed to launch editor")
		return
	}
	p.Refresh()
}

func seedPluginFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := data.EnsureDirPath(path, data.DefaultDirMod); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(pluginSeedYAML), data.DefaultFileMod)
}

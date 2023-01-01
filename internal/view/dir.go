package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell/v2"
)

const (
	kustomize      = "kustomization"
	kustomizeNoExt = "Kustomization"
	kustomizeYAML  = kustomize + extYAML
	kustomizeYML   = kustomize + extYML
	extYAML        = ".yaml"
	extYML         = ".yml"
)

// Dir represents a command directory view.
type Dir struct {
	ResourceViewer
	path string
}

// NewDir returns a new instance.
func NewDir(path string) ResourceViewer {
	d := Dir{
		ResourceViewer: NewBrowser(client.NewGVR("dir")),
		path:           path,
	}
	d.GetTable().SetBorderFocusColor(tcell.ColorAliceBlue)
	d.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorAliceBlue).Attributes(tcell.AttrNone))
	d.AddBindKeysFn(d.bindKeys)
	d.SetContextFn(d.dirContext)

	return &d
}

// Init initializes the view.
func (d *Dir) Init(ctx context.Context) error {
	if err := d.ResourceViewer.Init(ctx); err != nil {
		return err
	}

	return nil
}

func (d *Dir) dirContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPath, d.path)
}

func (d *Dir) bindDangerousKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyA: ui.NewKeyAction("Apply", d.applyCmd, true),
		ui.KeyD: ui.NewKeyAction("Delete", d.delCmd, true),
		ui.KeyE: ui.NewKeyAction("Edit", d.editCmd, true),
	})
}

func (d *Dir) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)
	if !d.App().Config.K9s.IsReadOnly() {
		d.bindDangerousKeys(aa)
	}
	aa.Add(ui.KeyActions{
		ui.KeyY:        ui.NewKeyAction("YAML", d.viewCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Goto", d.gotoCmd, true),
	})
}

func (d *Dir) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	if path.Ext(sel) == "" {
		return nil
	}

	yaml, err := os.ReadFile(sel)
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	details := NewDetails(d.App(), "YAML", sel, true).Update(string(yaml))
	if err := d.App().inject(details); err != nil {
		d.App().Flash().Err(err)
	}

	return nil
}

func isManifest(s string) bool {
	ext := path.Ext(s)
	return ext == ".yml" || ext == ".yaml"
}

func (d *Dir) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	if !isManifest(sel) {
		d.App().Flash().Errf("you must select a manifest")
		return nil
	}

	d.Stop()
	defer d.Start()
	if !edit(d.App(), shellOpts{clear: true, args: []string{sel}}) {
		d.App().Flash().Err(errors.New("Failed to launch editor"))
	}

	return nil
}

func (d *Dir) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.GetTable().CmdBuff().IsActive() {
		return d.GetTable().activateCmd(evt)
	}

	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	if isManifest(sel) {
		d.App().Flash().Errf("you must select a directory")
		return nil
	}

	v := NewDir(sel)
	if err := d.App().inject(v); err != nil {
		d.App().Flash().Err(err)
	}

	return evt
}

func isKustomized(sel string) bool {
	if isManifest(sel) {
		return false
	}

	ff, err := os.ReadDir(sel)
	if err != nil {
		return false
	}
	kk := []string{kustomizeNoExt, kustomizeYAML, kustomizeYML}
	for _, f := range ff {
		if config.InList(kk, f.Name()) {
			return true
		}
	}

	return false
}

func containsDir(sel string) bool {
	if isManifest(sel) {
		return false
	}

	ff, err := os.ReadDir(sel)
	if err != nil {
		return false
	}
	for _, f := range ff {
		if f.IsDir() {
			return true
		}
	}

	return false
}

func (d *Dir) applyCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	opts := []string{"-f"}
	if containsDir(sel) {
		opts = append(opts, "-R")
	}
	if isKustomized(sel) {
		opts = []string{"-k"}
	}
	d.Stop()
	defer d.Start()
	{
		args := make([]string, 0, 10)
		args = append(args, "apply")
		args = append(args, opts...)
		args = append(args, sel)
		res, err := runKu(d.App(), shellOpts{clear: false, args: args})
		if err != nil {
			res = "status:\n  " + err.Error() + "\nmessage:\n" + fmtResults(res)
		} else {
			res = "message:\n" + fmtResults(res)
		}

		details := NewDetails(d.App(), "Applied Manifest", sel, true).Update(res)
		if err := d.App().inject(details); err != nil {
			d.App().Flash().Err(err)
		}
	}

	return nil
}

func (d *Dir) delCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	d.Stop()
	defer d.Start()
	msg := fmt.Sprintf("Delete resource(s) in manifest %s", sel)
	dialog.ShowConfirm(d.App().Styles.Dialog(), d.App().Content.Pages, "Confirm Delete", msg, func() {
		args := make([]string, 0, 10)
		args = append(args, "delete")
		args = append(args, "-f")
		args = append(args, sel)
		res, err := runKu(d.App(), shellOpts{clear: false, args: args})
		if err != nil {
			res = "status:\n  " + err.Error() + "\nmessage:\n" + fmtResults(res)
		} else {
			res = "message:\n" + fmtResults(res)
		}
		details := NewDetails(d.App(), "Deleted Manifest", sel, true).Update(res)
		if err := d.App().inject(details); err != nil {
			d.App().Flash().Err(err)
		}
	}, func() {})

	return nil
}

func fmtResults(res string) string {
	res = strings.TrimSpace(res)
	lines := strings.Split(res, "\n")
	ll := make([]string, 0, len(lines))
	for _, l := range lines {
		ll = append(ll, "  "+l)
	}
	return strings.Join(ll, "\n")
}

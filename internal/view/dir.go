package view

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
	d.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorAliceBlue, tcell.AttrNone)
	d.SetBindKeysFn(d.bindKeys)
	d.SetContextFn(d.dirContext)
	d.GetTable().SetColorerFn(render.Dir{}.ColorerFunc())

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

func (d *Dir) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)
	aa.Add(ui.KeyActions{
		ui.KeyA:        ui.NewKeyAction("Apply", d.applyCmd, true),
		ui.KeyD:        ui.NewKeyAction("Delete", d.delCmd, true),
		ui.KeyE:        ui.NewKeyAction("Edit", d.editCmd, true),
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

	yaml, err := ioutil.ReadFile(sel)
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

	log.Debug().Msgf("Selected %q", sel)
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

func (d *Dir) applyCmd(evt *tcell.EventKey) *tcell.EventKey {
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
	{
		args := make([]string, 0, 10)
		args = append(args, "apply")
		args = append(args, "-f")
		args = append(args, sel)
		res, err := runKu(d.App(), shellOpts{clear: false, args: args})
		if err != nil {
			res = "error:\n  " + err.Error()
		} else {
			res = "message:\n  " + res
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

	if !isManifest(sel) {
		d.App().Flash().Errf("you must select a manifest")
		return nil
	}

	d.Stop()
	defer d.Start()
	msg := fmt.Sprintf("Delete resource(s) in manifest %s", sel)
	dialog.ShowConfirm(d.App().Content.Pages, "Confirm Delete", msg, func() {
		args := make([]string, 0, 10)
		args = append(args, "delete")
		args = append(args, "-f")
		args = append(args, sel)
		res, err := runKu(d.App(), shellOpts{clear: false, args: args})
		if err != nil {
			res = "error:\n  " + err.Error() + "\nmessage:\n  " + res
		} else {
			res = "message:\n  " + res
		}
		details := NewDetails(d.App(), "Deleted Manifest", sel, true).Update(res)
		if err := d.App().inject(details); err != nil {
			d.App().Flash().Err(err)
		}
	}, func() {})

	return nil
}

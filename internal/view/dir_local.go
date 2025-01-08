// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

// Dir represents a command directory view.
type DirLocal struct {
	ResourceViewer
	path       string
	fqn        string
	remote_dir string
}

// NewDirLocal returns a new instance.
func NewDirLocal(path string, fqn string, remote_dir string) ResourceViewer {
	d := DirLocal{
		ResourceViewer: NewBrowser(client.NewGVR("dirlocal")),
		path:           path,
		fqn:            fqn,
		remote_dir:     remote_dir,
	}
	d.GetTable().SetBorderFocusColor(tcell.ColorAliceBlue)
	d.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorAliceBlue).Attributes(tcell.AttrNone))
	d.AddBindKeysFn(d.bindKeys)
	d.SetContextFn(d.dirContext)

	return &d
}

// Init initializes the view.
func (d *DirLocal) Init(ctx context.Context) error {
	if err := d.ResourceViewer.Init(ctx); err != nil {
		return err
	}

	return nil
}

func (d *DirLocal) dirContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPath, d.path)
}

func (d *DirLocal) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyT: ui.NewKeyActionWithOpts("Transfer", d.transferCmd, ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		}),
	})
}

func (d *DirLocal) bindKeys(aa *ui.KeyActions) {
	// !!BOZO!! Lame!
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)
	if !d.App().Config.K9s.IsReadOnly() {
		d.bindDangerousKeys(aa)
	}
	aa.Bulk(ui.KeyMap{
		ui.KeyV:        ui.NewKeyAction("View", d.viewCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Goto", d.gotoCmd, true),
	})
}

func (d *DirLocal) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
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

	details := NewDetails(d.App(), yamlAction, sel, contentYAML, true).Update(string(yaml))
	if err := d.App().inject(details, false); err != nil {
		d.App().Flash().Err(err)
	}

	return nil
}

func (d *DirLocal) transferCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	path := d.fqn
	ns, n := client.Namespaced(path)
	file := d.GetTable().GetSelectedItem()

	pod, err := fetchPod(d.App().factory, path)
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	opts := dialog.TransferDialogOpts{
		Title:      "Transfer",
		Containers: fetchContainers(pod.ObjectMeta, pod.Spec, false),
		Message:    "Upload Files",
		Pod:        fmt.Sprintf("%s/%s:%s/%s", ns, n, d.remote_dir, filepath.Base(file)),
		File:       file,
		Ack:        makeTransferAck(d.App(), path),
		Download:   false,
		Retries:    defaultTxRetries,
		Cancel:     func() {},
	}
	dialog.ShowUploads(d.App().Styles.Dialog(), d.App().Content.Pages, opts)

	return nil
}

func (d *DirLocal) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.GetTable().CmdBuff().IsActive() {
		return d.GetTable().activateCmd(evt)
	}

	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	s, err := os.Stat(sel)
	if errors.Is(err, fs.ErrNotExist) {
		d.App().Flash().Err(err)
		return nil
	}

	if !s.IsDir() { // file
		return d.transferCmd(evt)
	}

	/*
		v := NewDirLocal(sel, d.fqn, d.remote_dir)
		if err := d.App().inject(v, false); err != nil {
			d.App().Flash().Err(err)
		}
	*/
	d.path = sel
	d.Start()

	return evt
}

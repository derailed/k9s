// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// Dir represents a command directory view.
type DirRemote struct {
	ResourceViewer
	fqn  string
	co   string
	os   string
	dir  string
	text string
}

// NewDirRemote returns a new instance.
func NewDirRemote(fqn, co string, os string, dir string, txt string) ResourceViewer {

	d := DirRemote{
		ResourceViewer: NewBrowser(client.NewGVR("dirremote")),
		fqn:            fqn,
		co:             co,
		os:             os,
		dir:            dir,
		text:           txt,
	}
	d.GetTable().SetBorderFocusColor(tcell.ColorAliceBlue)
	d.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorAliceBlue).Attributes(tcell.AttrNone))
	d.AddBindKeysFn(d.bindKeys)
	d.SetContextFn(d.dirContext)

	return &d
}

// Init initializes the view.
func (d *DirRemote) Init(ctx context.Context) error {
	if err := d.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	return nil
}

func (d *DirRemote) dirContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyContents, d.text)
	return context.WithValue(ctx, internal.KeyPath, d.dir)
}

func (d *DirRemote) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyT: ui.NewKeyActionWithOpts("Transfer", d.transferCmd, ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		}),
	})
}

func (d *DirRemote) bindKeys(aa *ui.KeyActions) {
	// !!BOZO!! Lame!
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)
	if !d.App().Config.K9s.IsReadOnly() {
		d.bindDangerousKeys(aa)
	}
	// TODO CD command
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Goto", d.gotoCmd, true),
		ui.KeyL:        ui.NewKeyAction("Local", d.localCmd, true),
	})
}

func getCWD(a *App, fqn, co string, os string) (string, error) {

	args := buildShellArgs("exec", fqn, co, a.Conn().Config().Flags().KubeConfig)
	if os == windowsOS {
		// FIXME implement
		// args = append(args, "--", powerShell)
	}

	// TODO make configurable
	args = append(args, "--", "sh", "-c", "echo $PWD")
	res, err := runKu(a, shellOpts{args: args})
	if err != nil {
		a.Flash().Errf("Shell exec '%s' failed: %s", strings.Join(args, " "), err)
	}

	return res, err
}

func listRemoteFiles(a *App, fqn, co string, os string, dir string) (string, error) {

	args := buildShellArgs("exec", fqn, co, a.Conn().Config().Flags().KubeConfig)
	if os == windowsOS {
		// FIXME implement
		// args = append(args, "--", powerShell)
	}

	// TODO make configurable
	args = append(args, "--", "ls", "-1aF", "--color=never")
	if dir != "" {
		args = append(args, "--", dir)
	}

	res, err := runKu(a, shellOpts{args: args})
	if err != nil {
		a.Flash().Errf("Shell exec '%s' failed: %s", strings.Join(args, " "), err)
	}

	return res, err
}

func (d *DirRemote) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.GetTable().CmdBuff().IsActive() {
		return d.GetTable().activateCmd(evt)
	}

	sel := d.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}
	if !strings.HasSuffix(sel, "/") { // file
		return d.transferCmd(evt)
	}
	dir := strings.TrimSuffix(sel, "/")

	str, err := listRemoteFiles(d.App(), d.fqn, d.co, d.os, dir)
	if err != nil {
		//d.App().Flash().Err(err)
		return evt
	}

	/*
		v := NewDirRemote(d.fqn, d.co, d.os, dir, str)
		if err := d.App().inject(v, false); err != nil {
			d.App().Flash().Err(err)
		}
	*/
	d.text = str
	d.dir = dir
	d.Start()

	return evt
}

func makeTransferAck(a *App, path string) func(args dialog.TransferArgs) bool {
	return func(args dialog.TransferArgs) bool {
		local := args.To
		if !args.Download {
			local = args.From
		}
		if _, err := os.Stat(local); !args.Download && errors.Is(err, fs.ErrNotExist) {
			a.Flash().Err(err)
			return false
		}

		opts := make([]string, 0, 10)
		opts = append(opts, "cp")
		opts = append(opts, strings.TrimSpace(args.From))
		opts = append(opts, strings.TrimSpace(args.To))
		opts = append(opts, fmt.Sprintf("--no-preserve=%t", args.NoPreserve))
		opts = append(opts, fmt.Sprintf("--retries=%d", args.Retries))
		if args.CO != "" {
			opts = append(opts, "-c="+args.CO)
		}
		opts = append(opts, fmt.Sprintf("--retries=%d", args.Retries))

		cliOpts := shellOpts{
			background: true,
			args:       opts,
		}
		op := trUpload
		if args.Download {
			op = trDownload
		}

		fqn := path + ":" + args.CO
		if err := runK(a, cliOpts); err != nil {
			a.cowCmd(err.Error())
		} else {
			a.Flash().Infof("%s successful on %s!", op, fqn)
		}
		return true
	}
}

func (d *DirRemote) transferCmd(evt *tcell.EventKey) *tcell.EventKey {
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
		Message:    "Download Files",
		Pod:        fmt.Sprintf("%s/%s:%s", ns, n, file),
		File:       filepath.Base(file),
		Ack:        makeTransferAck(d.App(), path),
		Download:   true,
		Retries:    defaultTxRetries,
		Cancel:     func() {},
	}
	dialog.ShowUploads(d.App().Styles.Dialog(), d.App().Content.Pages, opts)

	return nil
}

func (d *DirRemote) localCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := "."

	log.Debug().Msgf("DIR PATH %q", path)
	_, err := os.Stat(path)
	if err != nil {
		d.App().cowCmd(err.Error())
		return nil
	}
	if path == "." {
		dir, err := os.Getwd()
		if err == nil {
			path = dir
		}
	}
	d.App().cmdHistory.Push("dir " + path)

	if d.GetTable().CmdBuff().IsActive() {
		return d.GetTable().activateCmd(evt)
	}

	v := NewDirLocal(path, d.fqn, d.dir)
	if err := d.App().inject(v, false); err != nil {
		d.App().Flash().Err(err)
	}

	return evt
}

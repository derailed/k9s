// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

const containerFsTitle = "Container FS"

// ContainerFs represents a container filesystem browser view.
type ContainerFs struct {
	ResourceViewer
	podPath       string // e.g., "default/nginx-pod"
	containerName string // e.g., "nginx"
	currentDir    string // Current directory path
}

// NewContainerFs returns a new container filesystem browser.
func NewContainerFs(podPath, container, currentDir string) ResourceViewer {
	if currentDir == "" {
		currentDir = "/"
	}

	cf := ContainerFs{
		ResourceViewer: NewBrowser(client.CfsGVR),
		podPath:        podPath,
		containerName:  container,
		currentDir:     currentDir,
	}
	cf.GetTable().SetBorderFocusColor(tcell.ColorDodgerBlue)
	cf.GetTable().SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDodgerBlue).
		Attributes(tcell.AttrNone))
	cf.AddBindKeysFn(cf.bindKeys)
	cf.SetContextFn(cf.fsContext)

	return &cf
}

// Init initializes the view.
func (cf *ContainerFs) Init(ctx context.Context) error {
	if err := cf.ResourceViewer.Init(ctx); err != nil {
		return err
	}

	return nil
}

// Name returns the component name with directory path.
func (cf *ContainerFs) Name() string {
	// Show the directory path in the breadcrumb/title like Dir view does
	return fmt.Sprintf("%s:%s", containerFsTitle, cf.currentDir)
}

func (cf *ContainerFs) fsContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyPath, cf.podPath)
	ctx = context.WithValue(ctx, internal.KeyContainers, cf.containerName)
	return context.WithValue(ctx, internal.KeyCurrentDir, cf.currentDir)
}

func (cf *ContainerFs) bindKeys(aa *ui.KeyActions) {
	// Remove some default keys that don't apply here
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)

	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Goto", cf.gotoCmd, true),
	})
	// Note: Esc goes back by popping from the stack (built-in behavior)
}

// gotoCmd navigates into a directory or shows a file (mimics Dir view).
func (cf *ContainerFs) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := cf.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Check if it's a directory by looking at the NAME field icon
	// The render layer adds 📁 for directories and 📄 for files
	row := cf.GetTable().GetSelectedRow(sel)
	if row != nil && len(row.Fields) > 0 {
		name := row.Fields[0] // NAME column is first
		if !strings.HasPrefix(name, "📁 ") {
			// It's a file, show message for now (file viewing not implemented yet)
			cf.App().Flash().Infof("File viewing not yet implemented: %s", sel)
			return nil
		}
	}

	// Create new view for the selected directory and inject it
	v := NewContainerFs(cf.podPath, cf.containerName, sel)
	if err := cf.App().inject(v, false); err != nil {
		cf.App().Flash().Err(err)
	}

	return evt
}

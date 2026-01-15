// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"path/filepath"

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

	// Set title to show pod/container and current path
	title := fmt.Sprintf("%s [%s|%s:%s]",
		containerFsTitle, cf.podPath, cf.containerName, cf.currentDir)
	cf.GetTable().SetTitle(title)

	return nil
}

// Name returns the component name.
func (*ContainerFs) Name() string { return containerFsTitle }

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
		tcell.KeyEnter:  ui.NewKeyAction("Open", cf.openCmd, true),
		tcell.KeyRight:  ui.NewKeyAction("Open", cf.openCmd, true),
		tcell.KeyLeft:   ui.NewKeyAction("Back", cf.backCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Back", cf.backCmd, true),
	})
}

// Open directory or file.
func (cf *ContainerFs) openCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := cf.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Check if it's a directory by checking if path starts with folder icon
	row := cf.GetTable().GetSelectedRow(sel)
	if len(row.Fields) == 0 {
		return evt
	}

	// Check first character of NAME column - if it's the folder icon, it's a directory
	name := row.Fields[0]
	if len(name) < 4 || name[0:4] != "📁 " {
		// It's a file, show message for now (file viewing not implemented yet)
		cf.App().Flash().Infof("File viewing not yet implemented: %s", sel)
		return nil
	}

	// Navigate to subdirectory
	v := NewContainerFs(cf.podPath, cf.containerName, sel)
	if err := cf.App().inject(v, false); err != nil {
		cf.App().Flash().Err(err)
	}

	return nil
}

// Go up to parent directory.
func (cf *ContainerFs) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Already at root?
	if cf.currentDir == "/" {
		return evt
	}

	// Get parent directory
	parent := filepath.Dir(cf.currentDir)
	if parent == "." {
		parent = "/"
	}

	// Navigate to parent
	v := NewContainerFs(cf.podPath, cf.containerName, parent)
	if err := cf.App().inject(v, false); err != nil {
		cf.App().Flash().Err(err)
	}

	return nil
}

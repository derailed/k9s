// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
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
		ui.KeyG:        ui.NewSharedKeyAction("Goto Path", cf.activatePathCmd, false),
		ui.KeyD:        ui.NewKeyAction("Download", cf.downloadCmd, true),
	})
	// Note: Esc goes back by popping from the stack (built-in behavior)
}

// activatePathCmd shows a dialog for path navigation.
func (cf *ContainerFs) activatePathCmd(evt *tcell.EventKey) *tcell.EventKey {
	cf.showGotoPrompt()
	return nil
}

// gotoCmd navigates into a directory when Enter is pressed on selected item.
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

// showGotoPrompt shows a dialog for entering a path to navigate to.
func (cf *ContainerFs) showGotoPrompt() {
	styles := cf.App().Styles.Dialog()
	pages := cf.App().Content.Pages

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	var pathInput string
	f.AddInputField("Path:", "", 50, nil, func(text string) {
		pathInput = text
	})

	f.AddButton("Go", func() {
		pages.RemovePage("goto-prompt")
		if pathInput == "" {
			return
		}

		// Resolve the path (absolute or relative)
		targetPath := cf.resolvePath(pathInput)

		// Navigate to the target path
		v := NewContainerFs(cf.podPath, cf.containerName, targetPath)
		if err := cf.App().inject(v, false); err != nil {
			cf.App().Flash().Err(err)
		}
	})

	f.AddButton("Cancel", func() {
		pages.RemovePage("goto-prompt")
	})

	for i := range 2 {
		b := f.GetButton(i)
		if b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	modal := tview.NewModalForm("<Go to Path>", f)
	modal.SetText(fmt.Sprintf("Current: %s\nEnter absolute (/path) or relative (path) path:", cf.currentDir))
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		pages.RemovePage("goto-prompt")
	})

	pages.AddPage("goto-prompt", modal, false, false)
	pages.ShowPage("goto-prompt")
}

// resolvePath resolves a path (absolute or relative to currentDir).
func (cf *ContainerFs) resolvePath(input string) string {
	// If it starts with /, it's absolute
	if strings.HasPrefix(input, "/") {
		return input
	}

	// Otherwise, it's relative to currentDir
	// Clean up path separators
	if cf.currentDir == "/" {
		return "/" + input
	}
	return cf.currentDir + "/" + input
}

// downloadCmd downloads a file or directory from the container.
func (cf *ContainerFs) downloadCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := cf.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Check if it's a directory
	row := cf.GetTable().GetSelectedRow(sel)
	isDir := false
	var fileName string
	if row != nil && len(row.Fields) > 0 {
		name := row.Fields[0] // NAME column is first
		if strings.HasPrefix(name, "📁 ") {
			isDir = true
			fileName = strings.TrimPrefix(name, "📁 ")
		} else {
			fileName = strings.TrimPrefix(name, "📄 ")
		}
	}

	// Prompt for local save path
	defaultPath := "./" + fileName
	cf.showDownloadPrompt(sel, fileName, defaultPath, isDir)

	return nil
}

// showDownloadPrompt shows a simple input dialog for download path.
func (cf *ContainerFs) showDownloadPrompt(remotePath, fileName, defaultPath string, isDir bool) {
	styles := cf.App().Styles.Dialog()
	pages := cf.App().Content.Pages

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	var savePath string = defaultPath
	f.AddInputField("Save to:", defaultPath, 50, nil, func(text string) {
		savePath = text
	})

	f.AddButton("OK", func() {
		pages.RemovePage("download-prompt")
		if savePath == "" {
			cf.App().Flash().Warn("Download cancelled")
			return
		}

		// Download in background
		go func() {
			cf.App().Flash().Infof("Downloading %s to %s...", remotePath, savePath)
			var err error
			if isDir {
				err = cf.downloadDirectory(remotePath, savePath)
			} else {
				err = cf.downloadFile(remotePath, savePath)
			}

			if err != nil {
				cf.App().QueueUpdateDraw(func() {
					cf.App().Flash().Errf("Download failed: %s", err)
				})
			} else {
				cf.App().QueueUpdateDraw(func() {
					cf.App().Flash().Infof("Downloaded %s successfully!", fileName)
				})
			}
		}()
	})

	f.AddButton("Cancel", func() {
		pages.RemovePage("download-prompt")
	})

	for i := range 2 {
		b := f.GetButton(i)
		if b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	title := "<Download File>"
	itemType := "file"
	if isDir {
		title = "<Download Directory>"
		itemType = "directory"
	}

	modal := tview.NewModalForm(title, f)
	modal.SetText(fmt.Sprintf("Download %s: %s", itemType, fileName))
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		pages.RemovePage("download-prompt")
	})

	pages.AddPage("download-prompt", modal, false, false)
	pages.ShowPage("download-prompt")
}

// downloadFile downloads a file from the container.
func (cf *ContainerFs) downloadFile(remotePath, localPath string) error {
	// Get DAO
	var cfs dao.ContainerFs
	cfs.Init(cf.App().factory, client.CfsGVR)

	// Download the file
	ctx := context.Background()
	return cfs.DownloadFile(ctx, cf.podPath, cf.containerName, remotePath, localPath)
}

// downloadDirectory downloads a directory from the container.
func (cf *ContainerFs) downloadDirectory(remotePath, localPath string) error {
	// Get DAO
	var cfs dao.ContainerFs
	cfs.Init(cf.App().factory, client.CfsGVR)

	// Download the directory
	ctx := context.Background()
	return cfs.DownloadDirectory(ctx, cf.podPath, cf.containerName, remotePath, localPath)
}

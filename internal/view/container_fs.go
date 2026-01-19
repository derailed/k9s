// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const containerFsTitle = "Container FS"

// ContainerFs represents a container filesystem browser view.
type ContainerFs struct {
	ResourceViewer
	podPath       string          // e.g., "default/nginx-pod"
	containerName string          // e.g., "nginx"
	currentDir    string          // Current directory path
	pathBuff      *model.FishBuff // Buffer for interactive path navigation
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
		pathBuff:       model.NewFishBuff('/', model.FilterBuffer),
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

	// Wire up path navigation buffer
	cf.App().Prompt().SetModel(cf.pathBuff)
	cf.pathBuff.AddListener(cf)
	cf.pathBuff.SetSuggestionFn(cf.pathSuggestions)

	return nil
}

// Name returns the component name with directory path.
func (cf *ContainerFs) Name() string {
	// Show the directory path in the breadcrumb/title like Dir view does
	return fmt.Sprintf("%s:%s", containerFsTitle, cf.currentDir)
}

func (cf *ContainerFs) fsContext(ctx context.Context) context.Context {
	// KeyFQN stores the pod path for internal use (e.g., "default/nginx-pod")
	ctx = context.WithValue(ctx, internal.KeyFQN, cf.podPath)
	// KeyPath stores the descriptive display path (e.g., "default/nginx-pod:nginx:/var/log")
	displayPath := fmt.Sprintf("%s:%s:%s", cf.podPath, cf.containerName, cf.currentDir)
	ctx = context.WithValue(ctx, internal.KeyPath, displayPath)
	ctx = context.WithValue(ctx, internal.KeyContainers, cf.containerName)
	return context.WithValue(ctx, internal.KeyCurrentDir, cf.currentDir)
}

func (cf *ContainerFs) bindKeys(aa *ui.KeyActions) {
	// Remove some default keys that don't apply here
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, tcell.KeyCtrlZ)

	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Goto", cf.gotoCmd, true),
		tcell.KeyRight: ui.NewKeyAction("Goto", cf.gotoCmd, true),
		tcell.KeyLeft:  ui.NewKeyAction("Back", cf.App().PrevCmd, true),
		ui.KeyN:        ui.NewKeyAction("Navigate", cf.activatePathCmd, false),
		ui.KeyV:        ui.NewKeyAction("View", cf.viewFileCmd, true),
		ui.KeyD:        ui.NewKeyAction("Download", cf.downloadCmd, true),
	})
	// Note: Esc goes back by popping from the stack (built-in behavior)
}

// activatePathCmd activates the interactive path navigation prompt.
func (cf *ContainerFs) activatePathCmd(evt *tcell.EventKey) *tcell.EventKey {
	if cf.App().InCmdMode() {
		return evt
	}
	cf.App().ResetPrompt(cf.pathBuff)
	return nil
}

// gotoCmd navigates into a directory or views a file when Enter is pressed on selected item.
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
			// It's a file, view it
			cf.viewFile(sel)
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

// resolvePath resolves a path (absolute or relative to currentDir).
func (cf *ContainerFs) resolvePath(input string) string {
	if input == "" {
		return cf.currentDir
	}

	// Absolute path
	if strings.HasPrefix(input, "/") {
		return filepath.Clean(input)
	}

	// Relative path - join with current directory and clean
	resolved := filepath.Join(cf.currentDir, input)
	return filepath.Clean(resolved)
}

// pathSuggestions generates path completion suggestions based on current input.
func (cf *ContainerFs) pathSuggestions(text string) sort.StringSlice {
	var suggestions sort.StringSlice

	// Determine base directory and partial name for completion
	baseDir, partial := cf.parsePathForCompletion(text)

	// Get directory listing (uses cache if available)
	// If the directory doesn't exist, listSubdirectories will return an error
	// and we'll just return an empty suggestion list
	dirs, err := cf.listSubdirectories(baseDir)
	if err != nil {
		// Silently return empty suggestions if directory doesn't exist
		return suggestions
	}

	// Filter directories with prefix matching and return suffix to complete
	for _, dir := range dirs {
		if strings.HasPrefix(dir, partial) {
			// Return the suffix that should be added to complete the input
			// If partial is "da" and dir is "data", return "ta"
			suffix := strings.TrimPrefix(dir, partial)
			// Optionally add trailing slash for directories
			if suffix != "" {
				suggestions = append(suggestions, suffix+"/")
			} else if partial == dir {
				// If it's an exact match, suggest the slash
				suggestions = append(suggestions, "/")
			}
		}
	}

	suggestions.Sort()
	return suggestions
}

// parsePathForCompletion parses the input text to determine base directory and partial name.
func (cf *ContainerFs) parsePathForCompletion(text string) (baseDir, partial string) {
	if text == "" {
		return cf.currentDir, ""
	}

	// Resolve to absolute path first
	absPath := cf.resolvePath(text)

	// Check if path ends with / (user wants to see contents of that directory)
	if strings.HasSuffix(text, "/") {
		return absPath, ""
	}

	// Split into directory and partial filename
	dir, file := filepath.Split(absPath)
	if dir == "" {
		dir = "/"
	}
	// Clean trailing slash
	dir = strings.TrimSuffix(dir, "/")
	if dir == "" {
		dir = "/"
	}

	return dir, file
}

// listSubdirectories lists subdirectories of the given path from cache or DAO.
func (cf *ContainerFs) listSubdirectories(path string) ([]string, error) {
	// Use DAO to get listings (leverages cache)
	var cfs dao.ContainerFs
	cfs.Init(cf.App().factory, client.CfsGVR)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set context with pod, container, and directory
	ctx = context.WithValue(ctx, internal.KeyFQN, cf.podPath)
	ctx = context.WithValue(ctx, internal.KeyContainers, cf.containerName)
	ctx = context.WithValue(ctx, internal.KeyCurrentDir, path)

	// Fetch directory listing
	objs, err := cfs.List(ctx, "")
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, obj := range objs {
		// Type assert to ContainerFsRes
		if entry, ok := obj.(render.ContainerFsRes); ok {
			// Only include directories
			if entry.IsDir {
				dirs = append(dirs, entry.Name)
			}
		}
	}

	return dirs, nil
}

// BufferChanged is called on every keystroke (optional validation).
func (cf *ContainerFs) BufferChanged(text, suggestion string) {
	// Optional: could add real-time path validation feedback here
}

// BufferCompleted is called after typing pauses (debounced) and when Enter is pressed.
func (cf *ContainerFs) BufferCompleted(text, suggestion string) {
	if text == "" {
		return
	}

	// Only navigate if the buffer is no longer active
	// This means Enter was pressed (prompt deactivates buffer after Notify)
	// vs debounce timer (buffer still active)
	// We need to check this asynchronously since the prompt deactivates after calling Notify
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to let prompt deactivate
		if !cf.pathBuff.IsActive() {
			// Buffer was deactivated, meaning Enter was pressed
			targetPath := cf.resolvePath(text)

			// Navigate to the target path
			cf.App().QueueUpdateDraw(func() {
				v := NewContainerFs(cf.podPath, cf.containerName, targetPath)
				if err := cf.App().inject(v, false); err != nil {
					cf.App().Flash().Err(err)
				}
			})
		}
	}()
}

// BufferActive is called when buffer activation state changes.
func (cf *ContainerFs) BufferActive(state bool, k model.BufferKind) {
	cf.App().BufferActive(state, k)
}

// viewFileCmd handles the 'v' key to view a file.
func (cf *ContainerFs) viewFileCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := cf.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Check if it's a file
	row := cf.GetTable().GetSelectedRow(sel)
	if row != nil && len(row.Fields) > 0 {
		name := row.Fields[0] // NAME column is first
		if strings.HasPrefix(name, "📁 ") {
			cf.App().Flash().Warn("Cannot view directory, use Enter to navigate")
			return nil
		}
	}

	cf.viewFile(sel)
	return nil
}

// viewFile opens a file viewer for the given file path.
func (cf *ContainerFs) viewFile(filePath string) {
	// Show loading message
	cf.App().Flash().Infof("Loading file: %s", filePath)

	// Fetch file contents in background
	go func() {
		var cfs dao.ContainerFs
		cfs.Init(cf.App().factory, client.CfsGVR)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		content, err := cfs.ReadFile(ctx, cf.podPath, cf.containerName, filePath)
		if err != nil {
			cf.App().QueueUpdateDraw(func() {
				cf.App().Flash().Errf("Failed to read file: %s", err)
			})
			return
		}

		// Create and show file viewer
		cf.App().QueueUpdateDraw(func() {
			displayPath := fmt.Sprintf("%s:%s:%s", cf.podPath, cf.containerName, filePath)
			details := NewDetails(cf.App(), "File Viewer", displayPath, "text", true).Update(content)
			if err := cf.App().inject(details, false); err != nil {
				cf.App().Flash().Err(err)
			}
		})
	}()
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

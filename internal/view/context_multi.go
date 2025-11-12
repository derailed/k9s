// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tcell/v2"
)

// ContextExtended enhances Context view with multi-selection support.
type ContextExtended struct {
	*Context
}

// NewContextExtended returns a new context viewer with multi-selection support.
func NewContextExtended(gvr *client.GVR) ResourceViewer {
	c := ContextExtended{
		Context: &Context{
			ResourceViewer: NewBrowser(gvr),
		},
	}
	// Set enter function to use contexts
	c.GetTable().SetEnterFn(func(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
		c.useSelectedContextsCmd(nil)
	})
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *ContextExtended) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace)
	// Keep space key for marking multiple contexts
	aa.Add(ui.KeySpace, ui.NewKeyAction("Mark", c.markCmd, true))
	aa.Add(ui.KeyM, ui.NewKeyAction("Mark All", c.markAllCmd, true))
	aa.Add(ui.KeyU, ui.NewKeyAction("Unmark All", c.unmarkAllCmd, true))
	aa.Add(tcell.KeyEnter, ui.NewKeyAction("Use Context(s)", c.useSelectedContextsCmd, true))

	if !c.App().Config.IsReadOnly() {
		c.bindDangerousKeys(aa)
	}
}

func (c *ContextExtended) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyR, ui.NewKeyAction("Rename", c.renameCmd, true))
	aa.Add(tcell.KeyCtrlD, ui.NewKeyAction("Delete", c.deleteCmd, true))
}

func (c *ContextExtended) markCmd(evt *tcell.EventKey) *tcell.EventKey {
	c.GetTable().ToggleMark()
	return nil
}

func (c *ContextExtended) markAllCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Mark all contexts except the header
	table := c.GetTable()
	for i := 1; i < table.GetRowCount(); i++ {
		// Select this row and toggle mark
		table.Select(i, 0)
		table.ToggleMark()
	}
	c.Refresh()
	return nil
}

func (c *ContextExtended) unmarkAllCmd(evt *tcell.EventKey) *tcell.EventKey {
	c.GetTable().ClearMarks()
	c.Refresh()
	return nil
}

func (c *ContextExtended) useSelectedContextsCmd(evt *tcell.EventKey) *tcell.EventKey {
	selectedItems := c.GetTable().GetSelectedItems()
	if len(selectedItems) == 0 {
		return evt
	}

	// Sort contexts for consistent ordering
	sort.Strings(selectedItems)

	if len(selectedItems) == 1 {
		// Single context - use normal switch
		if err := useContext(c.App(), selectedItems[0]); err != nil {
			c.App().Flash().Err(err)
			return evt
		}
		c.App().clearHistory()
		c.Refresh()
		c.GetTable().Select(1, 0)
		return nil
	}

	// Multiple contexts - show confirmation and switch to multi-context mode
	contextNames := fmt.Sprintf("%v", selectedItems)
	d := c.App().Styles.Dialog()
	dialog.ShowConfirm(&d, c.App().Content.Pages,
		"Multi-Context Mode",
		fmt.Sprintf("Switch to multi-context mode with %d contexts?\n%s", len(selectedItems), contextNames),
		func() {
			if err := c.useMultipleContexts(selectedItems); err != nil {
				c.App().Flash().Err(err)
				return
			}
		},
		func() {})

	return nil
}

func (c *ContextExtended) useMultipleContexts(contexts []string) error {
	app := c.App()

	if app.Content.Top() != nil {
		app.Content.Top().Stop()
	}

	// Get the current config
	baseConfig := app.factory.Client().Config()

	// Create multi-connection
	slog.Info("Creating multi-context connection", slogs.Count, len(contexts))
	multiConn, err := client.NewMultiConnection(contexts, baseConfig, slog.Default())
	if err != nil {
		return fmt.Errorf("failed to create multi-connection: %w", err)
	}

	// Store old factory for cleanup
	oldFactory := app.factory

	// Switch to multi-context mode
	app.Config.K9s.ToggleContextSwitch(true)
	defer app.Config.K9s.ToggleContextSwitch(false)

	// Save config prior to context switch
	if err := app.Config.Save(true); err != nil {
		slog.Error("Failed to save config", slogs.Error, err)
	}

	// Update the app's connection
	app.Config.SetConnection(multiConn)

	// Stop old factory
	if oldFactory != nil {
		oldFactory.Terminate()
	}

	// Initialize new factory with multi-connection
	app.factory = watch.NewFactory(multiConn)
	app.initFactory(multiConn.ActiveNamespace())

	// Clear history and refresh
	app.clearHistory()

	// Flash success message
	if len(contexts) > 3 {
		app.Flash().Infof("Switched to multi-context mode with %d contexts", len(contexts))
	} else {
		app.Flash().Infof("Switched to multi-context mode: %v", contexts)
	}

	// Navigate to pods view to show the multi-context data
	app.gotoResource("pods", "", false, true)
	return nil
}

func (c *ContextExtended) renameCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	c.showRenameModal(contextName, c.renameDialogCallback)
	return nil
}

func (c *ContextExtended) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	d := c.App().Styles.Dialog()
	dialog.ShowConfirm(&d, c.App().Content.Pages, "Delete", fmt.Sprintf("Delete context %q?", contextName), func() {
		if err := c.App().factory.Client().Config().DelContext(contextName); err != nil {
			c.App().Flash().Err(err)
			return
		}
		c.Refresh()
	}, func() {})

	return nil
}

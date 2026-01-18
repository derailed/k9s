// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

const explainTitle = "Explain"

// Explain represents a kubectl explain view with tree navigation.
type Explain struct {
	ResourceViewer

	currentPath string
	app         *App
	pathHistory []string
	fields      []string
	fieldTypes  map[string]string
	fieldTree   []dao.FieldInfo
	kind        string
	version     string
	leafField   string
	leafType    string
	isLeaf      bool
	description string
}

// Stop stops the view.
func (e *Explain) Stop() {
	slog.Info("Explain Stop called", "path", e.currentPath)
	e.GetTable().Stop()
	e.GetTable().CmdBuff().RemoveListener(e)
}

// BufferChanged indicates the buffer was changed.
func (*Explain) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (*Explain) BufferCompleted(_, _ string) {}

// BufferActive indicates the buff activity changed.
func (*Explain) BufferActive(bool, model.BufferKind) {}

// NewExplain returns a new explain view.
func NewExplain(gvr *client.GVR) ResourceViewer {
	e := Explain{
		ResourceViewer: NewBrowser(gvr),
		pathHistory:    []string{},
		fields:         []string{},
	}
	e.AddBindKeysFn(e.bindKeys)

	return &e
}

// Init initializes the view.
func (e *Explain) Init(ctx context.Context) error {
	var err error
	if e.app, err = extractApp(ctx); err != nil {
		return err
	}

	if err := e.ResourceViewer.Init(ctx); err != nil {
		return err
	}

	e.GetTable().GetModel().SetNamespace(client.NotNamespaced)
	e.GetTable().SetTitle(" [aqua::b]Kubectl Explain[-::-] ")

	return nil
}

// Start starts the view.
func (e *Explain) Start() {
	// Don't call ResourceViewer.Start() as it tries to watch Kubernetes resources
	// Instead, just start the table and load our static explain data
	e.GetTable().Start()
	e.GetTable().CmdBuff().AddListener(e)

	// Load the explain content if a path was set
	if e.currentPath != "" {
		slog.Info("Explain Start: loading explain", "path", e.currentPath)
		e.loadExplain(e.currentPath)
	} else {
		slog.Info("Explain Start: no path set, showing help message")
		if e.app != nil {
			e.app.Flash().Info("Use :explain <resource> to explore Kubernetes resources (e.g., :explain pod, :explain deployment.spec)")
		}
		// Show a helpful table with examples
		e.showHelpTable()
	}
}

// SetInstance sets the current resource path to explain.
func (e *Explain) SetInstance(path string) {
	slog.Info("Explain SetInstance called", "path", path)
	e.currentPath = path
	e.pathHistory = []string{}
	e.fields = []string{}
	e.fieldTypes = make(map[string]string)
	e.fieldTree = []dao.FieldInfo{}
	e.kind = ""
	e.version = ""
	e.description = ""
	e.leafField = ""
	e.leafType = ""
	e.isLeaf = false
}

// bindKeys sets up key bindings.
func (e *Explain) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD, ui.KeyD, ui.KeyE)

	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter:  ui.NewKeyAction("Drill Down", e.drillDownCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Back", e.backCmd, true),
		ui.KeyR:         ui.NewKeyAction("Refresh", e.refreshCmd, true),
		tcell.KeyCtrlR:  ui.NewKeyAction("Refresh", e.refreshCmd, false),
		ui.KeyY:         ui.NewKeyAction("View Full", e.viewFullCmd, true),
	})
}

// drillDownCmd handles drilling down into a field.
func (e *Explain) drillDownCmd(evt *tcell.EventKey) *tcell.EventKey {
	if e.GetTable().CmdBuff().IsActive() {
		return e.GetTable().activateCmd(evt)
	}

	row, _ := e.GetTable().GetSelection()
	if row <= 0 || row > len(e.fields) {
		return nil
	}

	selectedField := e.fields[row-1]
	newPath := e.currentPath
	if newPath != "" {
		newPath += "."
	}
	newPath += selectedField

	// Save current path to history
	e.pathHistory = append(e.pathHistory, e.currentPath)

	// Load the new path
	e.loadExplain(newPath)

	return nil
}

// backCmd handles going back to the previous level.
func (e *Explain) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if len(e.pathHistory) == 0 {
		// If no history, reset to default view (pods) and go back
		// This prevents k9s from trying to restore explain view on restart
		e.app.Config.SetActiveView(client.PodGVR.String())
		return e.app.PrevCmd(evt)
	}

	// Pop the last path from history
	lastPath := e.pathHistory[len(e.pathHistory)-1]
	e.pathHistory = e.pathHistory[:len(e.pathHistory)-1]

	// Load the previous path
	e.loadExplain(lastPath)

	return nil
}

// refreshCmd refreshes the current view.
func (e *Explain) refreshCmd(evt *tcell.EventKey) *tcell.EventKey {
	e.loadExplain(e.currentPath)
	e.app.Flash().Info("Refreshed")
	return nil
}

// viewFullCmd shows the full recursive kubectl explain output for the selected field.
func (e *Explain) viewFullCmd(evt *tcell.EventKey) *tcell.EventKey {
	if e.currentPath == "" {
		return nil
	}

	// Get the selected field (one level deeper)
	row, _ := e.GetTable().GetSelection()
	if row <= 0 || row > len(e.fields) {
		e.app.Flash().Warn("No field selected")
		return nil
	}

	selectedField := e.fields[row-1]
	targetPath := e.currentPath
	if targetPath != "" {
		targetPath += "."
	}
	targetPath += selectedField

	e.app.Flash().Infof("Loading recursive explain for: %s", targetPath)

	ctx := context.WithValue(context.Background(), internal.KeyFactory, e.app.factory)
	explainDAO := &dao.Explain{}
	explainDAO.Init(e.app.factory, client.ExplainGVR)

	// Get recursive output for the selected field
	result, err := explainDAO.ExplainRecursive(ctx, targetPath)
	if err != nil {
		e.app.Flash().Errf("Failed to get recursive explain output: %v", err)
		return nil
	}

	// Show the full recursive content in a details view
	title := fmt.Sprintf("Explain (Recursive): %s", targetPath)
	details := NewDetails(e.app, title, targetPath, "text", true)
	details.Update(result.Content)
	if err := e.app.inject(details, false); err != nil {
		e.app.Flash().Err(err)
	}

	e.app.Flash().Infof("Showing recursive explain for: %s", targetPath)

	return nil
}

// loadExplain loads the explain content for the given path.
func (e *Explain) loadExplain(path string) {
	if path == "" {
		if e.app != nil {
			e.app.Flash().Warn("No resource path specified")
		}
		return
	}

	// Safety check - ensure app is initialized
	if e.app == nil {
		slog.Error("Explain view not properly initialized - app is nil")
		return
	}

	e.currentPath = path

	ctx := context.WithValue(context.Background(), internal.KeyFactory, e.app.factory)
	explainDAO := &dao.Explain{}
	explainDAO.Init(e.app.factory, client.ExplainGVR)

	result, err := explainDAO.Explain(ctx, path)
	if err != nil {
		e.app.Flash().Errf("Failed to explain %s: %v", path, err)
		slog.Error("Explain failed", slogs.Error, err, "path", path)
		return
	}

	slog.Info("Explain result received", "path", path, "fields_count", len(result.Fields), "content_length", len(result.Content), "is_leaf", result.IsLeaf)

	// Update fields, types, tree, kind, version, leaf info, and description
	e.fields = result.Fields
	e.fieldTypes = result.FieldTypes
	e.fieldTree = result.FieldTree
	e.kind = result.Kind
	e.version = result.Version
	e.leafField = result.LeafField
	e.leafType = result.LeafType
	e.description = result.Description
	e.isLeaf = result.IsLeaf

	slog.Info("About to update display", "fields", e.fields, "is_leaf", e.isLeaf)

	// For leaf nodes, show the full kubectl explain output as text
	if e.isLeaf {
		e.showLeafContent(result.Content)
		e.app.Flash().Infof("Leaf node: %s", path)
	} else {
		e.updateTable()
		e.app.Flash().Infof("Loaded explain for: %s (%d fields)", path, len(e.fields))
	}
}

// updateTable updates the table with current fields.
func (e *Explain) updateTable() {
	table := e.GetTable()

	slog.Info("updateTable called", "fields_count", len(e.fields), "fields", e.fields)

	// Update title with current path
	if e.currentPath != "" {
		table.SetTitle(fmt.Sprintf(" [aqua::b]Explain: %s[-::-] ", e.currentPath))
	} else {
		table.SetTitle(" [aqua::b]Kubectl Explain[-::-] ")
	}

	// Build header with field, type, version, and required status
	header := model1.Header{
		model1.HeaderColumn{Name: "FIELD"},
		model1.HeaderColumn{Name: "TYPE"},
		model1.HeaderColumn{Name: "VERSION"},
		model1.HeaderColumn{Name: "REQUIRED"},
	}

	// Build row events
	rowEvents := model1.NewRowEvents(len(e.fields))
	if len(e.fields) == 0 {
		slog.Info("No fields found (leaf node), showing single row with all details")

		// For leaf nodes, show everything in a single row
		fieldName := e.leafField
		if fieldName == "" {
			fieldName = "[Leaf Node]"
		}

		fieldType := e.leafType
		if fieldType == "" {
			fieldType = "<unknown>"
		}

		// Description in TYPE column (since it's the most important info for leaf nodes)
		description := e.description
		if description == "" {
			description = "No description available"
		}

		rowEvents.Add(model1.RowEvent{
			Kind: model1.EventUnchanged,
			Row: model1.Row{
				ID: "leaf-node",
				Fields: model1.Fields{
					e.kind,
					e.version,
					fieldName,
					fieldType + " - " + description,
					"No", // Leaf nodes are typically not required fields themselves
				},
			},
		})
	} else {
		// Use the recursive field tree if available
		if len(e.fieldTree) > 0 {
			slog.Info("Adding recursive field tree rows", "count", len(e.fieldTree))
			for i, fieldInfo := range e.fieldTree {
				// Build the tree representation with proper branch tracking
				var prefix strings.Builder

				// For each depth level, determine if we need | or space
				for d := 0; d < fieldInfo.Depth; d++ {
					// Check if there are more items at this depth level after current item
					hasMoreAtDepth := false
					for j := i + 1; j < len(e.fieldTree); j++ {
						if e.fieldTree[j].Depth < d {
							break
						}
						if e.fieldTree[j].Depth == d {
							hasMoreAtDepth = true
							break
						}
					}

					if hasMoreAtDepth {
						prefix.WriteString("│  ")
					} else {
						prefix.WriteString("   ")
					}
				}

				// Determine the branch character for this item
				isLastInBranch := false
				if i < len(e.fieldTree)-1 {
					nextDepth := e.fieldTree[i+1].Depth
					if nextDepth <= fieldInfo.Depth {
						isLastInBranch = true
					}
				} else {
					isLastInBranch = true
				}

				if isLastInBranch {
					prefix.WriteString("└─ ")
				} else {
					prefix.WriteString("├─ ")
				}

				treeField := prefix.String() + fieldInfo.Name

				// Build type string with required flag and enum
				typeStr := fieldInfo.Type
				if fieldInfo.Required {
					typeStr += " -required-"
				}

				if fieldInfo.EnumValues != "" {
					// Add enum on same line, truncated if too long
					enumStr := fieldInfo.EnumValues
					if len(enumStr) > 40 {
						enumStr = enumStr[:37] + "..."
					}
					typeStr += " -enum: " + enumStr
				}

				rowEvents.Add(model1.RowEvent{
					Kind: model1.EventUnchanged,
					Row: model1.Row{
						ID: fmt.Sprintf("field-%d", i),
						Fields: model1.Fields{
							treeField,
							typeStr,
						},
					},
				})
			}
		} else {
			// Simple field list (non-recursive)
			slog.Info("Adding simple field rows", "count", len(e.fields))
			for i, field := range e.fields {
				// Get the type for this field
				fieldType := e.fieldTypes[field]
				if fieldType == "" {
					fieldType = "<unknown>"
				}

				// Check if field is required (type contains -required-)
				required := "No"
				if strings.Contains(fieldType, "-required-") {
					required = "Yes"
					fieldType = strings.ReplaceAll(fieldType, "-required-", "")
					fieldType = strings.TrimSpace(fieldType)
				}

				rowEvents.Add(model1.RowEvent{
					Kind: model1.EventUnchanged,
					Row: model1.Row{
						ID: fmt.Sprintf("field-%d", i),
						Fields: model1.Fields{
							field,
							fieldType,
							e.version,
							required,
						},
					},
				})
			}
		}
	}

	// Create table data
	tableData := model1.NewTableDataFull(
		client.ExplainGVR,
		client.NotNamespaced,
		header,
		rowEvents,
	)

	slog.Info("Calling table.Update", "row_count", rowEvents.Len())

	// Update the table - must be done in QueueUpdateDraw for UI to render
	cdata := table.Update(tableData, false)
	e.app.QueueUpdateDraw(func() {
		table.UpdateUI(cdata, tableData)

		// Select first row if available
		if len(e.fields) > 0 {
			table.Select(1, 0)
		}

		slog.Info("updateTable UI updated")
	})

	slog.Info("updateTable completed")
}

// showLeafContent displays the full kubectl explain output for leaf nodes.
func (e *Explain) showLeafContent(content string) {
	// Create a Details view to show the text content
	details := NewDetails(e.app, fmt.Sprintf("Explain: %s", e.currentPath), e.currentPath, "text", true)
	details.Update(content)

	// Push the details view onto the stack
	if err := e.app.inject(details, false); err != nil {
		e.app.Flash().Err(err)
	}
}

// getTreePrefix returns the tree prefix for a field at the given index.
func (e *Explain) getTreePrefix(index, total int) string {
	if e.currentPath == "" {
		// Root level - no nesting
		if index == total-1 {
			return "└─ "
		}
		return "├─ "
	}

	// Split the path to determine depth
	parts := strings.Split(e.currentPath, ".")
	depth := len(parts)

	// Build the prefix based on depth
	var prefix strings.Builder

	// Add indentation for each level
	for i := 0; i < depth; i++ {
		if i < depth-1 {
			prefix.WriteString("│  ")
		} else {
			// Last level - add tree branch
			if index == total-1 {
				prefix.WriteString("└─ ")
			} else {
				prefix.WriteString("├─ ")
			}
		}
	}

	return prefix.String()
}

// getTreePath returns a visual tree representation of the current path (for leaf nodes).
func (e *Explain) getTreePath() string {
	if e.currentPath == "" {
		return ""
	}

	// Split the path into parts
	parts := strings.Split(e.currentPath, ".")
	if len(parts) == 0 {
		return ""
	}

	// Build tree visualization showing the full path
	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			result.WriteString("\n")
			result.WriteString(strings.Repeat("│  ", i-1))
			if i == len(parts)-1 {
				result.WriteString("└─ ")
			} else {
				result.WriteString("├─ ")
			}
		}
		result.WriteString(part)
	}

	return result.String()
}

// Name returns the view name.
func (e *Explain) Name() string {
	return explainTitle
}

// showHelpTable displays a helpful table with examples when no path is set.
func (e *Explain) showHelpTable() {
	header := model1.Header{
		model1.HeaderColumn{Name: "EXAMPLE COMMAND"},
		model1.HeaderColumn{Name: "DESCRIPTION"},
	}

	examples := []struct {
		command string
		desc    string
	}{
		{":explain pod", "Explore Pod resource structure"},
		{":explain deployment", "Explore Deployment resource structure"},
		{":explain service", "Explore Service resource structure"},
		{":explain pod.spec", "Explore Pod spec fields"},
		{":explain deployment.spec.template", "Explore Deployment template fields"},
		{":explain cronjob.spec.jobTemplate", "Explore CronJob job template"},
	}

	rowEvents := model1.NewRowEvents(len(examples))
	for i, ex := range examples {
		rowEvents.Add(model1.RowEvent{
			Kind: model1.EventUnchanged,
			Row: model1.Row{
				ID: fmt.Sprintf("example-%d", i),
				Fields: model1.Fields{
					ex.command,
					ex.desc,
				},
			},
		})
	}

	tableData := model1.NewTableDataFull(
		client.ExplainGVR,
		client.ClusterScope,
		header,
		rowEvents,
	)

	e.app.QueueUpdateDraw(func() {
		e.GetTable().SetTitle("Kubectl Explain - Examples")
		e.GetTable().UpdateUI(tableData, tableData)
	})
}

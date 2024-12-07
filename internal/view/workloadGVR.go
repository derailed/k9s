// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"gopkg.in/yaml.v3"
)

const workloadGVRTitle = "workloadGVR"

type WorkloadGVR struct {
	ResourceViewer
}

func NewWorkloadGVR(gvr client.GVR) ResourceViewer {
	a := WorkloadGVR{
		ResourceViewer: NewBrowser(gvr),
	}
	a.GetTable().SetBorderFocusColor(tcell.ColorAliceBlue)
	a.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorAliceBlue).Attributes(tcell.AttrNone))
	a.AddBindKeysFn(a.bindKeys)
	a.SetContextFn(a.workloadGVRContext)

	return &a
}

// Init initializes the view.
func (a *WorkloadGVR) Init(ctx context.Context) error {
	if err := a.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	a.GetTable().GetModel().SetNamespace(client.NotNamespaced)

	return nil
}

// workloadGVRContext initialise the worjloadGVR context
func (a *WorkloadGVR) workloadGVRContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyDir, a.App().Config.ContextWorkloadDir())
	return context.WithValue(ctx, internal.KeyPath, a.App().Config.ContextWorkloadPath())
}

func (a *WorkloadGVR) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyN, ui.KeyD, ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace, ui.KeyShiftD)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlD)
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftA: ui.NewKeyActionWithOpts("Context: add", a.addtoCurrentCtx,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyShiftD: ui.NewKeyActionWithOpts("Context: delete", a.deletefromCurrentCtx,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyC: ui.NewKeyActionWithOpts("Create custom GVR", a.createCustomCmd, ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		}),
		ui.KeyR: ui.NewKeyActionWithOpts("Delete custom GVR", a.deleteCustomCmd, ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		}),
		ui.KeyE: ui.NewKeyActionWithOpts("Edit custom GVR", a.editCustomCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyShiftG:   ui.NewKeyAction("Sort GVR", a.GetTable().SortColCmd("NAME", true), false),
		tcell.KeyEnter: ui.NewKeyAction("Simulate", a.simulateCmd, true),
		ui.KeyD:        ui.NewKeyAction("Show", a.describeCmd, true),
	})
}

// describeCmd will show a custom GVR
func (a *WorkloadGVR) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := a.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Retrieve custom workload gvr filepath
	pathFile := path.Join(a.App().Config.ContextWorkloadDir(), sel)
	data, err := os.ReadFile(pathFile)
	if err != nil {
		a.App().Flash().Err(err)
		return nil
	}

	// Describe custom workload GVR
	details := NewDetails(a.App(), "Describe", pathFile, contentYAML, true).Update(string(data))
	if err := a.App().inject(details, false); err != nil {
		a.App().Flash().Err(err)
		return nil
	}

	return nil
}

// createCustomCmd will create a custom worklad GVR wiht default template using a specified GVR's name
func (a *WorkloadGVR) createCustomCmd(evt *tcell.EventKey) *tcell.EventKey {
	var GVRName string

	// Generate creation form
	form, err := a.makeCreateForm(&GVRName)
	if err != nil {
		return nil
	}
	confirm := tview.NewModalForm("<Set GVR Name>", form)
	confirm.SetText(fmt.Sprintf("Set GVR Name %s %s", a.GVR(), a.App().Config.ContextWorkloadDir()))
	confirm.SetDoneFunc(func(int, string) {
		a.cleanupClusterContext()
		a.dismissDialog()
	})
	a.App().Content.AddPage("NewGVRModal", confirm, false, false)
	a.App().Content.ShowPage("NewGVRModal")

	return nil
}

func (a *WorkloadGVR) makeCreateForm(sel *string) (*tview.Form, error) {
	// Generate create form
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	f.AddInputField("GVR Filename", "", 0, nil, func(changed string) {
		*sel = changed
	})

	f.AddButton("OK", func() {
		defer a.dismissDialog()

		// Generate new filename / filepath
		filename := fmt.Sprintf("%s.yaml", *sel)
		filePathName := path.Join(a.App().Config.ContextWorkloadDir(), filename)

		// Create new GVR file
		if err := os.WriteFile(filePathName, config.Template, 0644); err != nil {
			a.App().Flash().Errf("Failed to create file: %q", err)
			return
		}

		a.Stop()
		defer a.Start()
		if !edit(a.App(), shellOpts{clear: true, args: []string{filePathName}}) {
			a.App().Flash().Err(errors.New("Failed to launch editor"))
			return
		}
	})
	f.AddButton("Cancel", func() {
		a.dismissDialog()
	})

	return f, nil
}

func (a *WorkloadGVR) dismissDialog() {
	a.App().Content.RemovePage("NewGVRModal")
}

// editCustomCmd will edit the current custom workloadGVR
func (a *WorkloadGVR) editCustomCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := a.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Edit existing custom GVR
	a.Stop()
	defer a.Start()
	if !edit(a.App(), shellOpts{clear: true, args: []string{path.Join(a.App().Config.ContextWorkloadDir(), sel)}}) {
		a.App().Flash().Err(errors.New("Failed to launch editor"))
		return nil
	}

	a.cleanupClusterContext()

	return nil
}

// deleteCustomCmd will delete the custom workload GVR
func (a *WorkloadGVR) deleteCustomCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := a.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Remove custom GRV (with prompt)
	filePath := path.Join(a.App().Config.ContextWorkloadDir(), sel)
	msg := fmt.Sprintf("Are you sure to delete the custom gvr: %s", strings.TrimSuffix(sel, filepath.Ext(sel)))
	dialog.ShowConfirm(a.App().Styles.Dialog(), a.App().Content.Pages, "Confirm Deletion", msg, func() {
		if err := os.Remove(filePath); err != nil {
			a.App().Flash().Errf("could not delete GVR: %q", err)
			return
		}

		a.cleanupClusterContext()
	}, func() {})

	return nil
}

// addtoCurrentCtx will add the GVR to the current cluster context
func (a *WorkloadGVR) addtoCurrentCtx(evt *tcell.EventKey) *tcell.EventKey {
	sel := a.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Add custom GVR to cluster context
	filenames := make([]string, 0)
	ctxWorkloadPath := a.App().Config.ContextWorkloadPath()
	content, err := os.ReadFile(ctxWorkloadPath)
	if err == nil {
		var ctxConfig config.WorkloadConfig
		if err := yaml.Unmarshal(content, &ctxConfig); err != nil {
			a.App().Flash().Errf("could not read workload context configuration: %q", err)
			return nil
		}
		filenames = ctxConfig.GVRFilenames
	}

	filenames = append(filenames, strings.TrimSuffix(sel, filepath.Ext(sel)))

	// Ensure there is no duplicate
	m := make(map[string]string)
	for _, n := range filenames {
		m[n] = n
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Save new config
	data, err := yaml.Marshal(config.WorkloadConfig{GVRFilenames: keys})
	if err != nil {
		a.App().Flash().Errf("could not marshal new configuration: %q", err)
		return nil
	}

	if err := os.WriteFile(ctxWorkloadPath, data, 0644); err != nil {
		a.App().Flash().Errf("could not write new configuration: %q", err)
		return nil
	}

	a.cleanupClusterContext()

	return nil
}

// deletefromCurrentCtx will delete the gvr from the current cluster context
func (a *WorkloadGVR) deletefromCurrentCtx(evt *tcell.EventKey) *tcell.EventKey {
	sel := a.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	// Delete custom GVR from cluster context
	filenames := make([]string, 0)
	ctxWorkloadPath := a.App().Config.ContextWorkloadPath()
	content, err := os.ReadFile(ctxWorkloadPath)
	if err == nil {
		var ctxConfig config.WorkloadConfig
		if err := yaml.Unmarshal(content, &ctxConfig); err != nil {
			a.App().Flash().Errf("could not unmarshal configuration: %q", err)
			return nil
		}
		filenames = ctxConfig.GVRFilenames
	}

	// Ensure there is no duplicate
	m := make(map[string]string)
	for _, n := range filenames {
		if n != strings.TrimSuffix(sel, filepath.Ext(sel)) {
			m[n] = n
		}
	}
	keys := make([]string, 0, len(m))
	for filename := range m {
		if _, err := os.Stat(fmt.Sprintf("%s.yaml", path.Join(a.App().Config.ContextWorkloadDir(), filename))); err == nil || !errors.Is(err, os.ErrNotExist) {
			keys = append(keys, filename)
		}
	}

	// Save new config
	data, err := yaml.Marshal(config.WorkloadConfig{GVRFilenames: keys})
	if err != nil {
		a.App().Flash().Errf("could not marshal new configuration: %q", err)
		return nil
	}

	if err := os.WriteFile(ctxWorkloadPath, data, 0644); err != nil {
		a.App().Flash().Errf("could not write new configuration: %q", err)
		return nil
	}

	a.cleanupClusterContext()

	return nil
}

func (a *WorkloadGVR) cleanupClusterContext() {
	validFilenames := make([]string, 0)

	// Get cluster context
	content, err := os.ReadFile(a.App().Config.ContextWorkloadPath())
	if err != nil {
		return
	}
	var ctxConfig config.WorkloadConfig
	if err := yaml.Unmarshal(content, &ctxConfig); err != nil {
		return
	}

	// Check if each files exists
	// Cleanup the one that doesn't exists
	for _, filename := range ctxConfig.GVRFilenames {
		if _, err := os.Stat(fmt.Sprintf("%s.yaml", path.Join(a.App().Config.ContextWorkloadDir(), filename))); err == nil || !errors.Is(err, os.ErrNotExist) {
			validFilenames = append(validFilenames, filename)
		}
	}

	// Save new cluster context
	data, err := yaml.Marshal(config.WorkloadConfig{GVRFilenames: validFilenames})
	if err != nil {
		a.App().Flash().Errf("could not marshal new configuration: %q", err)
		return
	}

	if err := os.WriteFile(a.App().Config.ContextWorkloadPath(), data, 0644); err != nil {
		a.App().Flash().Errf("could not write new configuration: %q", err)
		return
	}
}

// simulateCmd will show the custom workload GVR in the workload view
func (a *WorkloadGVR) simulateCmd(evt *tcell.EventKey) *tcell.EventKey {
	co := NewWorkload(client.NewGVR("workloads"))
	co.SetContextFn(a.singleWorkloadCtx)
	if err := a.App().inject(co, false); err != nil {
		a.App().Flash().Err(err)
		return nil
	}

	return evt
}

// singleWorkloadCtx will set the selected workloadGVR to the context to be simulated on the workload view
func (a *WorkloadGVR) singleWorkloadCtx(ctx context.Context) context.Context {
	wkgvrFilename := a.GetTable().GetSelectedItem()
	if wkgvrFilename == "" {
		return ctx
	}

	workloadcustomDir := a.App().Config.ContextWorkloadDir()
	wkgvr, err := config.GetWorkloadGVRFromFile(path.Join(workloadcustomDir, wkgvrFilename))
	if err != nil {
		a.App().Flash().Errf("could not retrieve workload gvr from file: %q", err)
		return ctx
	}

	return context.WithValue(ctx, internal.KeyWorkloadGVRs, []config.WorkloadGVR{wkgvr})
}

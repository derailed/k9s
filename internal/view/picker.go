// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// Picker represents a container picker.
type Picker struct {
	*tview.List

	actions ui.KeyActions
}

// NewPicker returns a new picker.
func NewPicker() *Picker {
	return &Picker{
		List:    tview.NewList(),
		actions: *ui.NewKeyActions(),
	}
}

func (p *Picker) SetFilter(string)                 {}
func (p *Picker) SetLabelFilter(map[string]string) {}

// Init initializes the view.
func (p *Picker) Init(ctx context.Context) error {
	app, err := extractApp(ctx)
	if err != nil {
		return err
	}

	pickerView := app.Styles.Views().Picker
	p.actions.Add(tcell.KeyEscape, ui.NewKeyAction("Back", app.PrevCmd, true))

	p.SetBorder(true)
	p.SetMainTextColor(pickerView.MainColor.Color())
	p.ShowSecondaryText(false)
	p.SetShortcutColor(pickerView.ShortcutColor.Color())
	p.SetSelectedBackgroundColor(pickerView.FocusColor.Color())
	p.SetTitle(" [aqua::b]Containers Picker ")

	p.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if a, ok := p.actions.Get(evt.Key()); ok {
			a.Action(evt)
			evt = nil
		}
		return evt
	})

	return nil
}

// InCmdMode checks if prompt is active.
func (*Picker) InCmdMode() bool {
	return false
}

// Start starts the view.
func (p *Picker) Start() {}

// Stop stops the view.
func (p *Picker) Stop() {}

// Name returns the component name.
func (p *Picker) Name() string { return "picker" }

// Hints returns the view hints.
func (p *Picker) Hints() model.MenuHints {
	return p.actions.Hints()
}

// ExtraHints returns additional hints.
func (p *Picker) ExtraHints() map[string]string {
	return nil
}

func (p *Picker) populate(ss []string) {
	p.Clear()
	for i, s := range ss {
		p.AddItem(s, "Select a container", rune('a'+i), nil)
	}
}

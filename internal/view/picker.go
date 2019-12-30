package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
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
		actions: ui.KeyActions{},
	}
}

// Init initializes the view.
func (v *Picker) Init(ctx context.Context) error {
	app, err := extractApp(ctx)
	if err != nil {
		return err
	}
	v.actions[tcell.KeyEscape] = ui.NewKeyAction("Back", app.PrevCmd, true)

	v.SetBorder(true)
	v.SetMainTextColor(tcell.ColorWhite)
	v.ShowSecondaryText(false)
	v.SetShortcutColor(tcell.ColorAqua)
	v.SetSelectedBackgroundColor(tcell.ColorAqua)
	v.SetTitle(" [aqua::b]Containers Picker ")
	v.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if a, ok := v.actions[evt.Key()]; ok {
			a.Action(evt)
			evt = nil
		}
		return evt
	})

	return nil
}

// Start starts the view.
func (v *Picker) Start() {}

// Stop stops the view.
func (v *Picker) Stop() {}

// Name returns the component name.
func (v *Picker) Name() string { return "picker" }

// Hints returns the view hints.
func (v *Picker) Hints() model.MenuHints {
	return v.actions.Hints()
}

func (v *Picker) populate(ss []string) {
	v.Clear()
	for i, s := range ss {
		v.AddItem(s, "Select a container", rune('a'+i), nil)
	}
}

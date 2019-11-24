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

func (v *Picker) Init(ctx context.Context) error {
	v.SetBorder(true)
	v.SetMainTextColor(tcell.ColorWhite)
	v.ShowSecondaryText(false)
	v.SetShortcutColor(tcell.ColorAqua)
	v.SetSelectedBackgroundColor(tcell.ColorAqua)
	v.SetTitle(" [aqua::b]Container Selector ")
	v.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if a, ok := v.actions[evt.Key()]; ok {
			a.Action(evt)
			evt = nil
		}
		return evt
	})

	return nil
}
func (v *Picker) Start()       {}
func (v *Picker) Stop()        {}
func (v *Picker) Name() string { return "picker" }

// Protocol...

func (v *Picker) Hints() model.MenuHints {
	return v.actions.Hints()
}

func (v *Picker) populate(ss []string) {
	v.Clear()
	for i, s := range ss {
		v.AddItem(s, "Select a container", rune('a'+i), nil)
	}
}

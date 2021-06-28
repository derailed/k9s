package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
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
func (p *Picker) Init(ctx context.Context) error {
	app, err := extractApp(ctx)
	if err != nil {
		return err
	}
	p.actions[tcell.KeyEscape] = ui.NewKeyAction("Back", app.PrevCmd, true)

	p.SetBorder(true)
	p.SetMainTextColor(tcell.ColorWhite)
	p.ShowSecondaryText(false)
	p.SetShortcutColor(tcell.ColorAqua)
	p.SetSelectedBackgroundColor(tcell.ColorAqua)
	p.SetTitle(" [aqua::b]Containers Picker ")
	p.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if a, ok := p.actions[evt.Key()]; ok {
			a.Action(evt)
			evt = nil
		}
		return evt
	})

	return nil
}

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

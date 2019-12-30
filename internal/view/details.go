package view

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// Details represents a generic text viewer.
type Details struct {
	*tview.TextView

	actions        ui.KeyActions
	app            *App
	title, subject string
	buff           string
}

// NewDetails returns a details viewer.
func NewDetails(title string) *Details {
	return &Details{
		TextView: tview.NewTextView(),
		title:    title,
		actions:  make(ui.KeyActions),
	}
}

// Init initializes the viewer.
func (d *Details) Init(ctx context.Context) error {
	var err error
	if d.app, err = extractApp(ctx); err != nil {
		return err
	}

	if d.title != "" {
		d.SetBorder(true)
	}
	d.SetBackgroundColor(d.app.Styles.BgColor())
	d.SetTextColor(d.app.Styles.FgColor())
	d.SetScrollable(true)
	d.SetWrap(true)
	d.SetDynamicColors(true)
	d.SetBorderFocusColor(config.AsColor(d.app.Styles.Frame().Border.FocusColor))
	d.SetHighlightColor(tcell.ColorOrange)
	d.SetTitleColor(tcell.ColorAqua)
	d.SetInputCapture(d.keyboard)
	d.bindKeys()
	d.SetChangedFunc(func() {
		d.app.Draw()
	})
	d.updateTitle()
	d.app.Styles.AddListener(d)

	return nil
}

func (d *Details) StylesChanged(s *config.Styles) {
	d.SetBackgroundColor(d.app.Styles.BgColor())
	d.SetTextColor(d.app.Styles.FgColor())
	d.Update(d.buff)
}

func (d *Details) Update(buff string) {
	d.buff = buff
	d.SetText(colorizeYAML(d.app.Styles.Views().Yaml, buff))
	d.ScrollToBeginning()
}

func (d *Details) Actions() ui.KeyActions {
	return d.actions
}

// Name returns the component name.
func (d *Details) Name() string { return d.title }

// Start starts the view updater.
func (d *Details) Start() {}

// Stop terminates the updater.
func (d *Details) Stop() {
	d.app.Styles.RemoveListener(d)
}

// Hints returns menu hints.
func (d *Details) Hints() model.MenuHints {
	return d.actions.Hints()
}

func (d *Details) bindKeys() {
	d.actions.Set(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", d.app.PrevCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", d.saveCmd, false),
		ui.KeyC:         ui.NewKeyAction("Copy", d.cpCmd, true),
	})
}

func (d *Details) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}

	if a, ok := d.actions[key]; ok {
		return a.Action(evt)
	}
	return evt
}

func (d *Details) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(d.app.Config.K9s.CurrentCluster, d.title, d.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	} else {
		d.app.Flash().Infof("Log %s saved successfully!", path)
	}
	return nil
}

func (d *Details) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	d.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(d.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	}
	return nil
}

func (d *Details) SetSubject(s string) {
	d.subject = s
}

func (d *Details) updateTitle() {
	if d.title == "" {
		return
	}
	title := ui.SkinTitle(fmt.Sprintf(detailsTitleFmt, d.title, d.subject), d.app.Styles.Frame())
	d.SetTitle(title)
}

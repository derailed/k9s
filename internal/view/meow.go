package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// Meow represents a bomb viewer
type Meow struct {
	*tview.TextView

	actions ui.KeyActions
	app     *App
	says    string
}

// NewMeow returns a details viewer.
func NewMeow(app *App, says string) *Meow {
	return &Meow{
		TextView: tview.NewTextView(),
		app:      app,
		actions:  make(ui.KeyActions),
		says:     says,
	}
}

// Init initializes the viewer.
func (m *Meow) Init(_ context.Context) error {
	m.SetBorder(true)
	m.SetScrollable(true).SetWrap(true).SetRegions(true)
	m.SetDynamicColors(true)
	m.SetHighlightColor(tcell.ColorOrange)
	m.SetTitleColor(tcell.ColorAqua)
	m.SetInputCapture(m.keyboard)
	m.SetBorderPadding(0, 0, 1, 1)
	m.updateTitle()
	m.SetTextAlign(tview.AlignCenter)

	m.app.Styles.AddListener(m)
	m.StylesChanged(m.app.Styles)

	m.bindKeys()
	m.SetInputCapture(m.keyboard)
	m.talk()

	return nil
}

func (m *Meow) talk() {
	says := m.says
	if len(says) == 0 {
		says = "Nothing to report here. Please move along..."
	}
	buff := make([]string, 0, len(cow)+3)
	buff = append(buff, " "+strings.Repeat("─", len(says)+8))
	buff = append(buff, fmt.Sprintf("< [red::b]MEOW! %s [-::-] >", says))
	buff = append(buff, " "+strings.Repeat("─", len(says)+8))
	spacer := strings.Repeat(" ", len(says)/2-8)
	for _, s := range cow {
		buff = append(buff, spacer+s)
	}
	m.SetText(strings.Join(buff, "\n"))
}

func (m *Meow) bindKeys() {
	m.actions.Set(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", m.resetCmd, false),
	})
}

func (m *Meow) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := m.actions[ui.AsKey(evt)]; ok {
		return a.Action(evt)
	}

	return evt
}

// StylesChanged notifies the skin changes.
func (m *Meow) StylesChanged(s *config.Styles) {
	m.SetBackgroundColor(m.app.Styles.BgColor())
	m.SetTextColor(m.app.Styles.FgColor())
	m.SetBorderFocusColor(m.app.Styles.Frame().Border.FocusColor.Color())
}

func (m *Meow) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	return m.app.PrevCmd(evt)
}

// Actions returns menu actions
func (m *Meow) Actions() ui.KeyActions {
	return m.actions
}

// Name returns the component name.
func (m *Meow) Name() string { return "cow" }

// Start starts the view updater.
func (m *Meow) Start() {}

// Stop terminates the updater.
func (m *Meow) Stop() {
	m.app.Styles.RemoveListener(m)
}

// Hints returns menu hints.
func (m *Meow) Hints() model.MenuHints {
	return m.actions.Hints()
}

// ExtraHints returns additional hints.
func (m *Meow) ExtraHints() map[string]string {
	return nil
}

func (m *Meow) updateTitle() {
	m.SetTitle(" Error ")
}

var cow = []string{
	`\   ^__^            `,
	` \  (oo)\_______    `,
	`    (__)\       )\/\`,
	`        ||----w |   `,
	`        ||     ||   `,
}

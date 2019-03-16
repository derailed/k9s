package views

import (
	"context"
	"fmt"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	helpTitle    = "Help"
	helpTitleFmt = " [aqua::b]%s "
)

type helpView struct {
	*tview.TextView

	app     *appView
	current igniter
	actions keyActions
}

func newHelpView(app *appView) *helpView {
	v := helpView{TextView: tview.NewTextView(), app: app, actions: make(keyActions)}
	{
		v.SetTextColor(tcell.ColorAqua)
		v.SetBorder(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
		v.SetInputCapture(v.keyboard)
		v.current = app.content.GetPrimitive("main").(igniter)
	}
	v.actions[tcell.KeyEsc] = newKeyAction("Back", v.backCmd, true)
	v.actions[tcell.KeyEnter] = newKeyAction("Back", v.backCmd, false)

	return &v
}

func (v *helpView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> TableView handled %s", tcell.KeyNames[key])
		return a.action(evt)
	}
	return evt
}

func (v *helpView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v.current)
	return nil
}

func (v *helpView) init(_ context.Context, _ string) {
	v.resetTitle()

	type helpItem struct {
		key, description string
	}

	general := []helpItem{
		{":<cmd>", "Command mode"},
		{"/<term>", "Filter mode"},
		{"esc", "Clear filter"},
		{"tab", "Next term match"},
		{"backtab", "Previous term match"},
		{"Ctrl-r", "Refresh"},
		{"p", "Previous resource view"},
		{"q", "Quit"},
	}
	fmt.Fprintf(v, "🏠 [aqua::b]%s\n", "General")
	for _, h := range general {
		fmt.Fprintf(v, "[pink::b]%9s [gray::]%s\n", h.key, h.description)
	}

	navigation := []helpItem{
		{"g", "Goto Top"},
		{"G", "Goto Bottom"},
		{"b", "Page Down"},
		{"f", "Page Up"},
		{"l", "Left"},
		{"h", "Right"},
		{"k", "Up"},
		{"j", "Down"},
	}
	fmt.Fprintf(v, "\n🤖 [aqua::b]%s\n", "View Navigation")
	for _, h := range navigation {
		fmt.Fprintf(v, "[pink::b]%9s [gray::]%s\n", h.key, h.description)
	}

	views := []helpItem{
		{"?", "Help"},
		{"a", "Aliases view"},
	}
	fmt.Fprintf(v, "️️\n😱 [aqua::b]%s\n", "Help")
	for _, h := range views {
		fmt.Fprintf(v, "[pink::b]%9s [gray::]%s\n", h.key, h.description)
	}

	v.app.setHints(v.hints())
}

func (v *helpView) hints() hints {
	return v.actions.toHints()
}

func (v *helpView) getTitle() string {
	return helpTitle
}

func (v *helpView) resetTitle() {
	v.SetTitle(fmt.Sprintf(helpTitleFmt, helpTitle))
}

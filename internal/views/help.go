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

type (
	helpItem struct {
		key, description string
	}

	helpView struct {
		*tview.TextView

		app     *appView
		current igniter
		actions keyActions
	}
)

func newHelpView(app *appView, current igniter) *helpView {
	v := helpView{
		TextView: tview.NewTextView(),
		app:      app,
		actions:  make(keyActions),
	}
	v.SetTextColor(tcell.ColorAqua)
	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetDynamicColors(true)
	v.SetInputCapture(v.keyboard)
	v.current = current
	v.bindKeys()

	return &v
}

func (v *helpView) bindKeys() {
	v.actions = keyActions{
		tcell.KeyEsc:   newKeyAction("Back", v.backCmd, true),
		tcell.KeyEnter: newKeyAction("Back", v.backCmd, false),
	}
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

	v.showGeneral()
	v.showNav()
	v.showHelp()
	v.app.setHints(v.hints())
}

func (v *helpView) showHelp() {
	views := []helpItem{
		{"?", "Help"},
		{"Ctrl-a", "Aliases view"},
	}
	fmt.Fprintf(v, "️️\n😱 [aqua::b]%s\n", "Help")
	for _, h := range views {
		v.printHelp(h.key, h.description)
	}
}

func (v *helpView) showNav() {
	navigation := []helpItem{
		{"g", "Goto Top"},
		{"G", "Goto Bottom"},
		{"Ctrl-b", "Page Down"},
		{"Ctrl-f", "Page Up"},
		{"h", "Left"},
		{"l", "Right"},
		{"k", "Up"},
		{"j", "Down"},
	}
	fmt.Fprintf(v, "\n🤖 [aqua::b]%s\n", "View Navigation")
	for _, h := range navigation {
		v.printHelp(h.key, h.description)
	}
}

func (v *helpView) showGeneral() {
	general := []helpItem{
		{":<cmd>", "Command mode"},
		{"/<term>", "Filter mode"},
		{"esc", "Clear filter"},
		{"tab", "Next term match"},
		{"backtab", "Previous term match"},
		{"Ctrl-r", "Refresh"},
		{"Shift-i", "Invert Sort"},
		{"p", "Previous resource view"},
		{":q", "Quit"},
	}
	fmt.Fprintf(v, "🏠 [aqua::b]%s\n", "General")
	for _, h := range general {
		v.printHelp(h.key, h.description)
	}
}

func (v *helpView) printHelp(key, desc string) {
	fmt.Fprintf(v, "[dodgerblue::b]%9s [white::]%s\n", key, desc)
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

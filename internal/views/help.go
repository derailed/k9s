package views

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/ui"
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
		*tview.Table

		app     *appView
		current ui.Igniter
		actions ui.KeyActions
	}
)

func newHelpView(app *appView, current ui.Igniter, hh ui.Hints) *helpView {
	v := helpView{
		Table:   tview.NewTable(),
		app:     app,
		actions: make(ui.KeyActions),
	}
	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetInputCapture(v.keyboard)
	v.current = current
	v.bindKeys()
	v.build(hh)

	return &v
}

func (v *helpView) bindKeys() {
	v.actions = ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", v.backCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Back", v.backCmd, false),
	}
}

func (v *helpView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> TableView handled %s", tcell.KeyNames[key])
		return a.Action(evt)
	}
	return evt
}

func (v *helpView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v.current)
	return nil
}

func (v *helpView) Init(_ context.Context, _ string) {
	v.resetTitle()
	v.app.SetHints(v.Hints())
}

func (v *helpView) showHelp() ui.Hints {
	return ui.Hints{
		{"?", "Help"},
		{"Ctrl-a", "Aliases view"},
	}
}

func (v *helpView) showNav() ui.Hints {
	return ui.Hints{
		{"g", "Goto Top"},
		{"G", "Goto Bottom"},
		{"Ctrl-b", "Page Down"},
		{"Ctrl-f", "Page Up"},
		{"h", "Left"},
		{"l", "Right"},
		{"k", "Up"},
		{"j", "Down"},
	}
}

func (v *helpView) showGeneral() ui.Hints {
	return ui.Hints{
		{":cmd", "Command mode"},
		{"/term", "Filter mode"},
		{"esc", "Clear filter"},
		{"tab", "Next term match"},
		{"backtab", "Previous term match"},
		{"Ctrl-r", "Refresh"},
		{"Shift-i", "Invert Sort"},
		{"p", "Previous resource view"},
		{":q", "Quit"},
	}
}

func (v *helpView) Hints() ui.Hints {
	return v.actions.Hints()
}

func (v *helpView) getTitle() string {
	return helpTitle
}

func (v *helpView) resetTitle() {
	v.SetTitle(fmt.Sprintf(helpTitleFmt, helpTitle))
}

func (v *helpView) build(hh ui.Hints) {
	v.Clear()
	sort.Sort(hh)
	v.addSection(0, 0, "Resource", hh)
	v.addSection(0, 4, "General", v.showGeneral())
	v.addSection(0, 6, "Navigation", v.showNav())
	v.addSection(0, 8, "Help", v.showHelp())
}

func (v *helpView) addSection(r, c int, title string, hh ui.Hints) {
	row := r
	cell := tview.NewTableCell(title)
	cell.SetTextColor(tcell.ColorWhite)
	cell.SetAttributes(tcell.AttrBold)
	v.SetCell(r, c, cell)
	row++

	for _, h := range hh {
		col := c
		cell := tview.NewTableCell(toMnemonic(h.Mnemonic))
		cell.SetTextColor(tcell.ColorDodgerBlue)
		cell.SetAttributes(tcell.AttrBold)
		cell.SetAlign(tview.AlignRight)
		v.SetCell(row, col, cell)
		col++
		cell = tview.NewTableCell(h.Description)
		cell.SetTextColor(tcell.ColorWhite)
		v.SetCell(row, col, cell)
		row++
	}
}

func toMnemonic(s string) string {
	if len(s) == 0 {
		return s
	}

	return "<" + keyConv(strings.ToLower(s)) + ">"
}

func keyConv(s string) string {
	if !strings.Contains(s, "alt") {
		return s
	}

	if runtime.GOOS != "darwin" {
		return s
	}

	return strings.Replace(s, "alt", "opt", 1)
}

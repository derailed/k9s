package view

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const (
	helpTitle    = "Help"
	helpTitleFmt = " [aqua::b]%s "
)

// Help presents a help viewer.
type Help struct {
	*Table

	maxKey, maxDesc, maxRows int
}

// NewHelp returns a new help viewer.
func NewHelp() *Help {
	return &Help{
		Table: NewTable(helpTitle),
	}
}

// Init initializes the component.
func (v *Help) Init(ctx context.Context) error {
	if err := v.Table.Init(ctx); err != nil {
		return nil
	}
	v.SetSelectable(false, false)
	v.resetTitle()
	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.bindKeys()
	v.build()
	v.SetBackgroundColor(v.App().Styles.BgColor())

	return nil
}

func (v *Help) bindKeys() {
	v.Actions().Delete(ui.KeySpace, tcell.KeyCtrlSpace, tcell.KeyCtrlS)
	v.Actions().Set(ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", v.app.PrevCmd, false),
		ui.KeyHelp:     ui.NewKeyAction("Back", v.app.PrevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Back", v.app.PrevCmd, false),
	})
}

func (v *Help) computeMaxes(hh model.MenuHints) {
	v.maxKey, v.maxDesc = 0, 0
	for _, h := range hh {
		if len(h.Mnemonic) > v.maxKey {
			v.maxKey = len(h.Mnemonic)
		}
		if len(h.Description) > v.maxDesc {
			v.maxDesc = len(h.Description)
		}
	}
	v.maxKey += 2
}

type HelpFunc func() model.MenuHints

func (v *Help) build() {
	v.Clear()

	ff := []HelpFunc{v.app.Content.Top().Hints, v.showGeneral, v.showNav, v.showHelp}
	var col int
	for i, section := range []string{"RESOURCE", "GENERAL", "NAVIGATION", "HELP"} {
		hh := ff[i]()
		sort.Sort(hh)
		v.computeMaxes(hh)
		v.addSection(col, section, hh)
		col += 2
	}

	if h, err := v.showHotKeys(); err == nil {
		v.computeMaxes(h)
		v.addSection(col, "HOTKEYS", h)
	}
}

func (v *Help) showHelp() model.MenuHints {
	return model.MenuHints{
		{
			Mnemonic:    "?",
			Description: "Help",
		},
		{
			Mnemonic:    "Ctrl-a",
			Description: "Aliases",
		},
	}
}

func (v *Help) showNav() model.MenuHints {
	return model.MenuHints{
		{
			Mnemonic:    "g",
			Description: "Goto Top",
		},
		{
			Mnemonic:    "Shift-g",
			Description: "Goto Bottom",
		},
		{
			Mnemonic:    "Ctrl-b",
			Description: "Page Down"},
		{
			Mnemonic:    "Ctrl-f",
			Description: "Page Up",
		},
		{
			Mnemonic:    "h",
			Description: "Left",
		},
		{
			Mnemonic:    "l",
			Description: "Right",
		},
		{
			Mnemonic:    "k",
			Description: "Up",
		},
		{
			Mnemonic:    "j",
			Description: "Down",
		},
	}
}

func (v *Help) showHotKeys() (model.MenuHints, error) {
	hh := config.NewHotKeys()
	if err := hh.Load(); err != nil {
		return nil, fmt.Errorf("no hotkey configuration found")
	}
	kk := make(sort.StringSlice, 0, len(hh.HotKey))
	for k := range hh.HotKey {
		kk = append(kk, k)
	}
	kk.Sort()
	mm := make(model.MenuHints, 0, len(hh.HotKey))
	for _, k := range kk {
		mm = append(mm, model.MenuHint{
			Mnemonic:    hh.HotKey[k].ShortCut,
			Description: hh.HotKey[k].Description,
		})
	}

	return mm, nil
}

func (v *Help) showGeneral() model.MenuHints {
	return model.MenuHints{
		{
			Mnemonic:    ":cmd",
			Description: "Command mode",
		},
		{
			Mnemonic:    "/term",
			Description: "Filter mode",
		},
		{
			Mnemonic:    "esc",
			Description: "Back/Clear",
		},
		{
			Mnemonic:    "tab",
			Description: "Next Field",
		},
		{
			Mnemonic:    "backtab",
			Description: "Previous Field",
		},
		{
			Mnemonic:    "Ctrl-r",
			Description: "Reload",
		},
		{
			Mnemonic:    "Ctrl-u",
			Description: "Clear command",
		},
		{
			Mnemonic:    "h",
			Description: "Toggle Header",
		},
		{
			Mnemonic:    ":q",
			Description: "Quit",
		},
		{
			Mnemonic:    "space",
			Description: "Mark",
		},
		{
			Mnemonic:    "Ctrl-space",
			Description: "Clear Marks",
		},
		{
			Mnemonic:    "Ctrl-s",
			Description: "Save",
		},
	}
}

func (v *Help) resetTitle() {
	v.SetTitle(fmt.Sprintf(helpTitleFmt, helpTitle))
}

func (v *Help) addSpacer(c int) {
	cell := tview.NewTableCell(render.Pad("", v.maxKey))
	cell.SetBackgroundColor(v.App().Styles.BgColor())
	cell.SetExpansion(1)
	v.SetCell(0, c, cell)
}

func (v *Help) addSection(c int, title string, hh model.MenuHints) {
	if len(hh) > v.maxRows {
		v.maxRows = len(hh)
	}
	row := 0
	cell := tview.NewTableCell(title)
	cell.SetTextColor(tcell.ColorGreen)
	cell.SetAttributes(tcell.AttrBold)
	cell.SetExpansion(1)
	cell.SetAlign(tview.AlignLeft)
	v.SetCell(row, c, cell)
	v.addSpacer(c + 1)
	row++

	for _, h := range hh {
		col := c
		cell := tview.NewTableCell(render.Pad(toMnemonic(h.Mnemonic), v.maxKey))
		if _, err := strconv.Atoi(h.Mnemonic); err != nil {
			cell.SetTextColor(tcell.ColorDodgerBlue)
		} else {
			cell.SetTextColor(tcell.ColorFuchsia)
		}
		cell.SetAttributes(tcell.AttrBold)
		v.SetCell(row, col, cell)
		col++
		cell = tview.NewTableCell(render.Pad(h.Description, v.maxDesc))
		cell.SetTextColor(tcell.ColorWhite)
		v.SetCell(row, col, cell)
		row++
	}

	if len(hh) < v.maxRows {
		for i := v.maxRows - len(hh); i > 0; i-- {
			col := c
			cell := tview.NewTableCell(render.Pad("", v.maxKey))
			v.SetCell(row, col, cell)
			col++
			cell = tview.NewTableCell(render.Pad("", v.maxDesc))
			v.SetCell(row, col, cell)
			row++
		}
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

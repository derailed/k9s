package view

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
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

// HelpFunc processes menu hints.
type HelpFunc func() model.MenuHints

// Help presents a help viewer.
type Help struct {
	*Table

	maxKey, maxDesc, maxRows int
}

// NewHelp returns a new help viewer.
func NewHelp() *Help {
	return &Help{
		Table: NewTable(client.NewGVR("help")),
	}
}

// Init initializes the component.
func (h *Help) Init(ctx context.Context) error {
	if err := h.Table.Init(ctx); err != nil {
		return nil
	}
	h.SetSelectable(false, false)
	h.resetTitle()
	h.SetBorder(true)
	h.SetBorderPadding(0, 0, 1, 1)
	h.bindKeys()
	h.build()
	h.SetBackgroundColor(h.App().Styles.BgColor())

	return nil
}

func (h *Help) bindKeys() {
	h.Actions().Delete(ui.KeySpace, tcell.KeyCtrlSpace, tcell.KeyCtrlS)
	h.Actions().Set(ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", h.app.PrevCmd, false),
		ui.KeyHelp:     ui.NewKeyAction("Back", h.app.PrevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Back", h.app.PrevCmd, false),
	})
}

func (h *Help) computeMaxes(hh model.MenuHints) {
	h.maxKey, h.maxDesc = 0, 0
	for _, hint := range hh {
		if len(hint.Mnemonic) > h.maxKey {
			h.maxKey = len(hint.Mnemonic)
		}
		if len(hint.Description) > h.maxDesc {
			h.maxDesc = len(hint.Description)
		}
	}
	h.maxKey += 2
}

func (h *Help) build() {
	h.Clear()

	h.maxRows = len(h.showGeneral())
	ff := []HelpFunc{h.app.Content.Top().Hints, h.showGeneral, h.showNav, h.showHelp}
	var col int
	for i, section := range []string{"RESOURCE", "GENERAL", "NAVIGATION", "HELP"} {
		hh := ff[i]()
		sort.Sort(hh)
		h.computeMaxes(hh)
		h.addSection(col, section, hh)
		col += 2
	}

	if hh, err := h.showHotKeys(); err == nil {
		h.computeMaxes(hh)
		h.addSection(col, "HOTKEYS", hh)
	}
}

func (h *Help) showHelp() model.MenuHints {
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

func (h *Help) showNav() model.MenuHints {
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

func (h *Help) showHotKeys() (model.MenuHints, error) {
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

func (h *Help) showGeneral() model.MenuHints {
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
			Mnemonic:    "Ctrl-h",
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

func (h *Help) resetTitle() {
	h.SetTitle(fmt.Sprintf(helpTitleFmt, helpTitle))
}

func (h *Help) addSpacer(c int) {
	cell := tview.NewTableCell(render.Pad("", h.maxKey))
	cell.SetBackgroundColor(h.App().Styles.BgColor())
	cell.SetExpansion(1)
	h.SetCell(0, c, cell)
}

func (h *Help) addSection(c int, title string, hh model.MenuHints) {
	if len(hh) > h.maxRows {
		h.maxRows = len(hh)
	}
	row := 0
	cell := tview.NewTableCell(title)
	cell.SetTextColor(tcell.ColorGreen)
	cell.SetAttributes(tcell.AttrBold)
	cell.SetExpansion(1)
	cell.SetAlign(tview.AlignLeft)
	h.SetCell(row, c, cell)
	h.addSpacer(c + 1)
	row++

	for _, hint := range hh {
		col := c
		cell := tview.NewTableCell(render.Pad(toMnemonic(hint.Mnemonic), h.maxKey))
		if _, err := strconv.Atoi(hint.Mnemonic); err != nil {
			cell.SetTextColor(tcell.ColorDodgerBlue)
		} else {
			cell.SetTextColor(tcell.ColorFuchsia)
		}
		cell.SetAttributes(tcell.AttrBold)
		h.SetCell(row, col, cell)
		col++
		cell = tview.NewTableCell(render.Pad(hint.Description, h.maxDesc))
		cell.SetTextColor(tcell.ColorWhite)
		h.SetCell(row, col, cell)
		row++
	}

	if len(hh) < h.maxRows {
		for i := h.maxRows - len(hh); i > 0; i-- {
			col := c
			cell := tview.NewTableCell(render.Pad("", h.maxKey))
			h.SetCell(row, col, cell)
			col++
			cell = tview.NewTableCell(render.Pad("", h.maxDesc))
			h.SetCell(row, col, cell)
			row++
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

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

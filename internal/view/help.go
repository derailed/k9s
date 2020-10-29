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
		tcell.KeyEscape: ui.NewKeyAction("Back", h.app.PrevCmd, true),
		ui.KeyHelp:      ui.NewKeyAction("Back", h.app.PrevCmd, false),
		tcell.KeyEnter:  ui.NewKeyAction("Back", h.app.PrevCmd, false),
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

func (h *Help) computeExtraMaxes(ee map[string]string) {
	h.maxDesc = 0
	for k := range ee {
		if len(k) > h.maxDesc {
			h.maxDesc = len(k)
		}
	}
}

func (h *Help) build() {
	h.Clear()

	sections := []string{"RESOURCE", "GENERAL", "NAVIGATION", "HELP"}

	h.maxRows = len(h.showGeneral())
	ff := []HelpFunc{
		h.app.Content.Top().Hints,
		h.showGeneral,
		h.showNav,
		h.showHelp,
	}
	var col int
	extras := h.app.Content.Top().ExtraHints()
	for i, section := range sections {
		hh := ff[i]()
		sort.Sort(hh)
		h.computeMaxes(hh)
		if extras != nil {
			h.computeExtraMaxes(extras)
		}
		h.addSection(col, section, hh)
		if i == 0 && extras != nil {
			h.addExtras(extras, col, len(hh))
		}
		col += 2
	}

	if hh, err := h.showHotKeys(); err == nil {
		h.computeMaxes(hh)
		h.addSection(col, "HOTKEYS", hh)
	}
}

func (h *Help) addExtras(extras map[string]string, col, size int) {
	kk := make([]string, 0, len(extras))
	for k := range extras {
		kk = append(kk, k)
	}
	sort.StringSlice(kk).Sort()
	row := size + 1
	for _, k := range kk {
		h.SetCell(row, col, padCell(extras[k], h.maxKey))
		h.SetCell(row, col+1, padCell(k, h.maxDesc))
		row++
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
			Description: "Page Up"},
		{
			Mnemonic:    "Ctrl-f",
			Description: "Page Down",
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
			Description: "Field Next",
		},
		{
			Mnemonic:    "backtab",
			Description: "Field Previous",
		},
		{
			Mnemonic:    "Ctrl-r",
			Description: "Reload",
		},
		{
			Mnemonic:    "Ctrl-u",
			Description: "Command Clear",
		},
		{
			Mnemonic:    "Ctrl-e",
			Description: "Toggle Header",
		},
		{
			Mnemonic:    "Ctrl-g",
			Description: "Toggle Crumbs",
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
			Description: "Mark Range",
		},
		{
			Mnemonic:    "Ctrl-\\",
			Description: "Mark Clear",
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
	h.SetCell(row, c, titleCell(title))
	h.addSpacer(c + 1)
	row++

	for _, hint := range hh {
		col := c
		h.SetCell(row, col, keyCell(hint.Mnemonic, h.maxKey))
		col++
		h.SetCell(row, col, infoCell(hint.Description, h.maxDesc))
		row++
	}

	if len(hh) >= h.maxRows {
		return
	}

	for i := h.maxRows - len(hh); i > 0; i-- {
		col := c
		h.SetCell(row, col, padCell("", h.maxKey))
		col++
		h.SetCell(row, col, padCell("", h.maxDesc))
		row++
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

func titleCell(title string) *tview.TableCell {
	c := tview.NewTableCell(title)
	c.SetTextColor(tcell.ColorGreen)
	c.SetAttributes(tcell.AttrBold)
	c.SetExpansion(1)
	c.SetAlign(tview.AlignLeft)

	return c
}

func keyCell(k string, width int) *tview.TableCell {
	c := padCell(toMnemonic(k), width)
	if _, err := strconv.Atoi(k); err != nil {
		c.SetTextColor(tcell.ColorDodgerBlue)
	} else {
		c.SetTextColor(tcell.ColorFuchsia)
	}
	c.SetAttributes(tcell.AttrBold)

	return c
}

func infoCell(info string, width int) *tview.TableCell {
	c := padCell(info, width)
	c.SetTextColor(tcell.ColorWhite)

	return c
}

func padCell(s string, width int) *tview.TableCell {
	return tview.NewTableCell(render.Pad(s, width))
}

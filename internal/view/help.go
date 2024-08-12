// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
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

	styles                   *config.Styles
	hints                    HelpFunc
	maxKey, maxDesc, maxRows int
}

// NewHelp returns a new help viewer.
func NewHelp(app *App) *Help {
	return &Help{
		Table: NewTable(client.NewGVR("help")),
		hints: app.Content.Top().Hints,
	}
}

func (h *Help) SetFilter(string)                 {}
func (h *Help) SetLabelFilter(map[string]string) {}

// Init initializes the component.
func (h *Help) Init(ctx context.Context) error {
	if err := h.Table.Init(ctx); err != nil {
		return err
	}
	h.SetSelectable(false, false)
	h.resetTitle()
	h.SetBorder(true)
	h.SetBorderPadding(0, 0, 1, 1)
	h.bindKeys()
	h.build()
	h.app.Styles.AddListener(h)
	h.StylesChanged(h.app.Styles)

	return nil
}

// InCmdMode checks if prompt is active.
func (*Help) InCmdMode() bool {
	return false
}

// StylesChanged notifies skin changed.
func (h *Help) StylesChanged(s *config.Styles) {
	h.styles = s
	h.SetBackgroundColor(s.BgColor())
	h.updateStyle()
}

func (h *Help) bindKeys() {
	h.Actions().Delete(ui.KeySpace, tcell.KeyCtrlSpace, tcell.KeyCtrlS, ui.KeySlash)
	h.Actions().Bulk(ui.KeyMap{
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

	sections := []string{"RESOURCE", "GENERAL", "NAVIGATION"}
	h.maxRows = len(h.showGeneral())
	ff := []HelpFunc{
		h.hints,
		h.showGeneral,
		h.showNav,
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
			Description: "Page Up",
		},
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
	if err := hh.Load(h.App().Config.ContextHotkeysPath()); err != nil {
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
			Mnemonic:    "?",
			Description: "Help",
		},
		{
			Mnemonic:    "Ctrl-a",
			Description: "Aliases",
		},
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
	cell.SetExpansion(1)
	h.SetCell(0, c, cell)
}

func (h *Help) addSection(c int, title string, hh model.MenuHints) {
	if len(hh) > h.maxRows {
		h.maxRows = len(hh)
	}
	row := 0
	h.SetCell(row, c, h.titleCell(title))
	h.addSpacer(c + 1)
	row++

	for _, hint := range hh {
		col := c
		h.SetCell(row, col, padCellWithRef(ui.ToMnemonic(hint.Mnemonic), h.maxKey, hint.Mnemonic))
		col++
		h.SetCell(row, col, padCell(hint.Description, h.maxDesc))
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

func (h *Help) updateStyle() {
	var (
		style   = tcell.StyleDefault.Background(h.styles.K9s.Help.BgColor.Color())
		key     = style.Foreground(h.styles.K9s.Help.KeyColor.Color()).Bold(true)
		numKey  = style.Foreground(h.app.Styles.K9s.Help.NumKeyColor.Color()).Bold(true)
		info    = style.Foreground(h.app.Styles.K9s.Help.FgColor.Color())
		heading = style.Foreground(h.app.Styles.K9s.Help.SectionColor.Color())
	)
	for col := 0; col < h.GetColumnCount(); col++ {
		for row := 0; row < h.GetRowCount(); row++ {
			c := h.GetCell(row, col)
			if c == nil {
				continue
			}
			switch {
			case row == 0:
				c.SetStyle(heading)
			case col%2 != 0:
				c.SetStyle(info)
			default:
				if _, err := strconv.Atoi(extractRef(c)); err == nil {
					c.SetStyle(numKey)
					continue
				}
				c.SetStyle(key)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func extractRef(c *tview.TableCell) string {
	if ref, ok := c.GetReference().(string); ok {
		return ref
	}

	return c.Text
}

func (h *Help) titleCell(title string) *tview.TableCell {
	c := tview.NewTableCell(title)
	c.SetTextColor(h.Styles().K9s.Help.SectionColor.Color())
	c.SetAttributes(tcell.AttrBold)
	c.SetExpansion(1)
	c.SetAlign(tview.AlignLeft)

	return c
}

func padCellWithRef(s string, width int, ref interface{}) *tview.TableCell {
	return padCell(s, width).SetReference(ref)
}

func padCell(s string, width int) *tview.TableCell {
	return tview.NewTableCell(render.Pad(s, width))
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	menuIndexFmt = " [key:-:b]<%d> [fg:-:d]%s "
	maxRows      = 6
)

var menuRX = regexp.MustCompile(`\d`)

// Menu presents menu options.
type Menu struct {
	*tview.Table

	styles *config.Styles
}

// NewMenu returns a new menu.
func NewMenu(styles *config.Styles) *Menu {
	m := Menu{
		Table:  tview.NewTable(),
		styles: styles,
	}
	m.SetBackgroundColor(styles.BgColor())
	styles.AddListener(&m)

	return &m
}

// StylesChanged notifies skin changed.
func (m *Menu) StylesChanged(s *config.Styles) {
	m.styles = s
	m.SetBackgroundColor(s.BgColor())
	for row := 0; row < m.GetRowCount(); row++ {
		for col := 0; col < m.GetColumnCount(); col++ {
			if c := m.GetCell(row, col); c != nil {
				c.BackgroundColor = s.BgColor()
			}
		}
	}
}

// StackPushed notifies a component was added.
func (m *Menu) StackPushed(c model.Component) {
	m.HydrateMenu(c.Hints())
}

// StackPopped notifies a component was removed.
func (m *Menu) StackPopped(o, top model.Component) {
	if top != nil {
		m.HydrateMenu(top.Hints())
	} else {
		m.Clear()
	}
}

// StackTop notifies the top component.
func (m *Menu) StackTop(t model.Component) {
	m.HydrateMenu(t.Hints())
}

// HydrateMenu populate menu ui from hints.
func (m *Menu) HydrateMenu(hh model.MenuHints) {
	m.Clear()
	sort.Sort(hh)

	table := make([]model.MenuHints, maxRows+1)
	colCount := (len(hh) / maxRows) + 1
	if m.hasDigits(hh) {
		colCount++
	}
	for row := 0; row < maxRows; row++ {
		table[row] = make(model.MenuHints, colCount)
	}
	t := m.buildMenuTable(hh, table, colCount)

	for row := 0; row < len(t); row++ {
		for col := 0; col < len(t[row]); col++ {
			c := tview.NewTableCell(t[row][col])
			if len(t[row][col]) == 0 {
				c = tview.NewTableCell("")
			}
			c.SetBackgroundColor(m.styles.BgColor())
			m.SetCell(row, col, c)
		}
	}
}

func (m *Menu) hasDigits(hh model.MenuHints) bool {
	for _, h := range hh {
		if !h.Visible {
			continue
		}
		if menuRX.MatchString(h.Mnemonic) {
			return true
		}
	}
	return false
}

func (m *Menu) buildMenuTable(hh model.MenuHints, table []model.MenuHints, colCount int) [][]string {
	var row, col int
	firstCmd := true
	maxKeys := make([]int, colCount)
	for _, h := range hh {
		if !h.Visible {
			continue
		}

		if !menuRX.MatchString(h.Mnemonic) && firstCmd {
			row, col, firstCmd = 0, col+1, false
			if table[0][0].IsBlank() {
				col = 0
			}
		}
		if maxKeys[col] < len(h.Mnemonic) {
			maxKeys[col] = len(h.Mnemonic)
		}
		table[row][col] = h
		row++
		if row >= maxRows {
			row, col = 0, col+1
		}
	}

	out := make([][]string, len(table))
	for r := range out {
		out[r] = make([]string, len(table[r]))
	}
	m.layout(table, maxKeys, out)

	return out
}

func (m *Menu) layout(table []model.MenuHints, mm []int, out [][]string) {
	for r := range table {
		for c := range table[r] {
			out[r][c] = m.formatMenu(table[r][c], mm[c])
		}
	}
}

func (m *Menu) formatMenu(h model.MenuHint, size int) string {
	if h.Mnemonic == "" || h.Description == "" {
		return ""
	}
	i, err := strconv.Atoi(h.Mnemonic)
	if err == nil {
		return formatNSMenu(i, h.Description, m.styles.Frame())
	}

	return formatPlainMenu(h, size, m.styles.Frame())
}

// ----------------------------------------------------------------------------
// Helpers...

func keyConv(s string) string {
	if s == "" || !strings.Contains(s, "alt") {
		return s
	}
	if runtime.GOOS != "darwin" {
		return s
	}

	return strings.Replace(s, "alt", "opt", 1)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

func ToMnemonic(s string) string {
	if s == "" {
		return s
	}

	return "<" + keyConv(strings.ToLower(s)) + ">"
}

func formatNSMenu(i int, name string, styles config.Frame) string {
	fmat := strings.Replace(menuIndexFmt, "[key", "["+styles.Menu.NumKeyColor.String(), 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+styles.Title.BgColor.String()+":", -1)
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor.String(), 1)

	return fmt.Sprintf(fmat, i, name)
}

func formatPlainMenu(h model.MenuHint, size int, styles config.Frame) string {
	menuFmt := " [key:-:b]%-" + strconv.Itoa(size+2) + "s [fg:-:d]%s "
	fmat := strings.Replace(menuFmt, "[key", "["+styles.Menu.KeyColor.String(), 1)
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor.String(), 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+styles.Title.BgColor.String()+":", -1)

	return fmt.Sprintf(fmat, ToMnemonic(h.Mnemonic), h.Description)
}

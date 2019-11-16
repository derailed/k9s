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
	menuIndexFmt = " [key:bg:b]<%d> [fg:bg:d]%s "
	maxRows      = 7
	chopWidth    = 20
)

var menuRX = regexp.MustCompile(`\d`)

// Menu presents menu options.
type Menu struct {
	*tview.Table

	styles *config.Styles
}

// NewMenu returns a new menu.
func NewMenu(styles *config.Styles) *Menu {
	v := Menu{Table: tview.NewTable(), styles: styles}
	v.SetBackgroundColor(styles.BgColor())

	return &v
}

func (v *Menu) StackPushed(c model.Component) {
	v.HydrateMenu(c.Hints())
}

func (v *Menu) StackPopped(o, top model.Component) {
	if top != nil {
		v.HydrateMenu(top.Hints())
	} else {
		v.Clear()
	}
}

func (v *Menu) StackTop(t model.Component) {
	v.HydrateMenu(t.Hints())
}

// HydrateMenu populate menu ui from hints.
func (v *Menu) HydrateMenu(hh model.MenuHints) {
	v.Clear()
	sort.Sort(hh)

	table := make([]model.MenuHints, maxRows+1)
	colCount := (len(hh) / maxRows) + 1
	if v.hasDigits(hh) {
		colCount++
	}
	for row := 0; row < maxRows; row++ {
		table[row] = make(model.MenuHints, colCount)
	}
	t := v.buildMenuTable(hh, table, colCount)

	for row := 0; row < len(t); row++ {
		for col := 0; col < len(t[row]); col++ {
			if len(t[row][col]) == 0 {
				continue
			}
			c := tview.NewTableCell(t[row][col])
			c.SetBackgroundColor(v.styles.BgColor())
			v.SetCell(row, col, c)
		}
	}
}

func (v *Menu) hasDigits(hh model.MenuHints) bool {
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

func (v *Menu) buildMenuTable(hh model.MenuHints, table []model.MenuHints, colCount int) [][]string {
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
	v.layout(table, maxKeys, out)

	return out
}

func (v *Menu) layout(table []model.MenuHints, mm []int, out [][]string) {
	for r := range table {
		for c := range table[r] {
			out[r][c] = keyConv(v.formatMenu(table[r][c], mm[c]))
		}
	}

}

func (v *Menu) formatMenu(h model.MenuHint, size int) string {
	if h.Mnemonic == "" || h.Description == "" {
		return ""
	}
	i, err := strconv.Atoi(h.Mnemonic)
	if err == nil {
		return formatNSMenu(i, h.Description, v.styles.Frame())
	}

	return formatPlainMenu(h, size, v.styles.Frame())
}

// ----------------------------------------------------------------------------
// Helpers...

func keyConv(s string) string {
	if !strings.Contains(s, "alt") {
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

func toMnemonic(s string) string {
	if len(s) == 0 {
		return s
	}

	return "<" + keyConv(strings.ToLower(s)) + ">"
}

func formatNSMenu(i int, name string, styles config.Frame) string {
	fmat := strings.Replace(menuIndexFmt, "[key", "["+styles.Menu.NumKeyColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+styles.Title.BgColor+":", -1)
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor, 1)
	return fmt.Sprintf(fmat, i, Truncate(name, chopWidth))
}

func formatPlainMenu(h model.MenuHint, size int, styles config.Frame) string {
	menuFmt := " [key:bg:b]%-" + strconv.Itoa(size+2) + "s [fg:bg:d]%s "
	fmat := strings.Replace(menuFmt, "[key", "["+styles.Menu.KeyColor, 1)
	fmat = strings.Replace(fmat, "[fg", "["+styles.Menu.FgColor, 1)
	fmat = strings.Replace(fmat, ":bg:", ":"+styles.Title.BgColor+":", -1)
	return fmt.Sprintf(fmat, toMnemonic(h.Mnemonic), h.Description)
}

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

// Help presents a help viewer.
type Help struct {
	*Table
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
	v.resetTitle()
	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.bindKeys()
	v.build(v.app.Content.Top().Hints())

	return nil
}

func (v *Help) bindKeys() {
	v.Actions().Delete(ui.KeySpace, tcell.KeyCtrlSpace, tcell.KeyCtrlS)
	v.Actions().Set(ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", v.app.PrevCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Back", v.app.PrevCmd, false),
	})
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
	m := make(model.MenuHints, 0, len(hh.HotKey))
	for _, hk := range hh.HotKey {
		m = append(m, model.MenuHint{
			Mnemonic:    hk.ShortCut,
			Description: hk.Description,
		})
	}

	return m, nil
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
			Description: "Refresh",
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

func (v *Help) build(hh model.MenuHints) {
	v.Clear()
	sort.Sort(hh)
	var col int
	v.addSection(col, "RESOURCE", hh)
	col += 2
	v.addSection(col, "GENERAL", v.showGeneral())
	col += 2
	v.addSection(col, "NAVIGATION", v.showNav())
	col += 2
	if h, err := v.showHotKeys(); err == nil {
		v.addSection(col, "HOTKEYS", h)
		col += 2
	}
	v.addSection(col, "HELP", v.showHelp())
}

func (v *Help) addSection(c int, title string, hh model.MenuHints) {
	row := 0
	cell := tview.NewTableCell(title)
	cell.SetTextColor(tcell.ColorGreen)
	cell.SetAttributes(tcell.AttrBold)
	cell.SetExpansion(2)
	cell.SetAlign(tview.AlignLeft)
	v.SetCell(row, c+1, cell)
	row++

	for _, h := range hh {
		col := c
		cell := tview.NewTableCell(toMnemonic(h.Mnemonic))
		if _, err := strconv.Atoi(h.Mnemonic); err != nil {
			cell.SetTextColor(tcell.ColorDodgerBlue)
		} else {
			cell.SetTextColor(tcell.ColorFuchsia)
		}
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

func defaultK9sEnv(app *App, sel string, row render.Row) K9sEnv {
	ns, n := client.Namespaced(sel)
	ctx, err := app.Conn().Config().CurrentContextName()
	if err != nil {
		ctx = render.NAValue
	}
	cluster, err := app.Conn().Config().CurrentClusterName()
	if err != nil {
		cluster = render.NAValue
	}
	user, err := app.Conn().Config().CurrentUserName()
	if err != nil {
		user = render.NAValue
	}
	groups, err := app.Conn().Config().CurrentGroupNames()
	if err != nil {
		groups = []string{render.NAValue}
	}
	var cfg string
	kcfg := app.Conn().Config().Flags().KubeConfig
	if kcfg != nil && *kcfg != "" {
		cfg = *kcfg
	}

	env := K9sEnv{
		"NAMESPACE":  ns,
		"NAME":       n,
		"CONTEXT":    ctx,
		"CLUSTER":    cluster,
		"USER":       user,
		"GROUPS":     strings.Join(groups, ","),
		"KUBECONFIG": cfg,
	}

	for i, r := range row.Fields {
		env["COL"+strconv.Itoa(i)] = r
	}

	return env
}

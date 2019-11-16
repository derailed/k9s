package view

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	helpTitle    = "Help"
	helpTitleFmt = " [aqua::b]%s "
)

// Help presents a help viewer.
type Help struct {
	*ui.Table

	app     *App
	actions ui.KeyActions
}

// NewHelp returns a new help viewer.
func NewHelp() *Help {
	return &Help{
		Table:   ui.NewTable(helpTitle),
		actions: make(ui.KeyActions),
	}
}

func (v *Help) Init(ctx context.Context) {
	v.app = mustExtractApp(ctx)

	v.resetTitle()

	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetInputCapture(v.keyboard)
	v.bindKeys()
	v.build(v.app.Content.Previous().Hints())
}

func (v *Help) Name() string { return helpTitle }
func (v *Help) Start()       {}
func (v *Help) Stop()        {}
func (v *Help) Hints() model.MenuHints {
	log.Debug().Msgf("Help Hints %#v", v.actions.Hints())
	return v.actions.Hints()
}

func (v *Help) bindKeys() {
	v.RmActions(tcell.KeyCtrlSpace, ui.KeySpace)

	v.actions = ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", v.backCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Back", v.backCmd, false),
	}
}

func (v *Help) keyboard(evt *tcell.EventKey) *tcell.EventKey {
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

func (v *Help) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	return v.app.PrevCmd(evt)
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
			Description: "Clear filter",
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
			Mnemonic:    "Shift-i",
			Description: "Invert Sort",
		},
		{
			Mnemonic:    ":q",
			Description: "Quit",
		},
	}
}

func (v *Help) resetTitle() {
	v.SetTitle(fmt.Sprintf(helpTitleFmt, helpTitle))
}

func (v *Help) build(hh model.MenuHints) {
	v.Clear()
	sort.Sort(hh)
	v.addSection(0, "RESOURCE", hh)
	v.addSection(4, "GENERAL", v.showGeneral())
	v.addSection(6, "NAVIGATION", v.showNav())
	v.addSection(8, "HELP", v.showHelp())
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

func defaultK9sEnv(app *App, sel string, row resource.Row) K9sEnv {
	ns, n := namespaced(sel)
	ctx, err := app.Conn().Config().CurrentContextName()
	if err != nil {
		ctx = resource.NAValue
	}
	cluster, err := app.Conn().Config().CurrentClusterName()
	if err != nil {
		cluster = resource.NAValue
	}
	user, err := app.Conn().Config().CurrentUserName()
	if err != nil {
		user = resource.NAValue
	}
	groups, err := app.Conn().Config().CurrentGroupNames()
	if err != nil {
		groups = []string{resource.NAValue}
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

	for i, r := range row {
		env["COL"+strconv.Itoa(i)] = r
	}

	return env
}

package views

import (
	"context"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

type helpView struct {
	*tview.Pages

	app     *appView
	title   string
	current igniter
	keys    keyActions
}

func newHelpView(app *appView) *helpView {
	return &helpView{
		app:     app,
		title:   "Help",
		current: app.content.GetPrimitive("main").(igniter),
		Pages:   tview.NewPages(),
	}
}

// Init the view.
func (v *helpView) init(context.Context, string) {
	v.keys = keyActions{
		tcell.KeyEscape: keyAction{description: "Back", action: v.back},
	}

	t := tview.NewTable()
	{
		t.SetBorder(true)
		t.SetTitle(" [::b]Commands Aliases ")
		t.SetTitleColor(tcell.ColorAqua)
		t.SetBorderColor(tcell.ColorDodgerBlue)
		t.SetSelectable(true, false)
		t.SetSelectedStyle(tcell.ColorWhite, tcell.ColorFuchsia, tcell.AttrNone)
		t.SetInputCapture(v.keyboard)
	}

	var row int
	for c, h := range []string{"ALIAS", "RESOURCE", "APIGROUP"} {
		th := tview.NewTableCell(h)
		th.SetExpansion(3)
		t.SetCell(row, c, th)
	}
	row++

	cmds := helpCmds()
	kk := make([]string, 0, len(cmds))
	for k := range cmds {
		kk = append(kk, k)
	}
	sort.Strings(kk)
	var col int
	for _, k := range kk {
		tc := tview.NewTableCell(resource.Pad(k, 30))
		tc.SetExpansion(2)
		t.SetCell(row, col, tc)
		col++
		tc = tview.NewTableCell(resource.Pad(cmds[k].title, 30))
		tc.SetExpansion(2)
		t.SetCell(row, col, tc)
		col++
		tc = tview.NewTableCell(resource.Pad(cmds[k].api, 30))
		tc.SetExpansion(2)
		t.SetCell(row, col, tc)
		col = 0
		row++
	}

	v.AddPage("main", t, true, true)
	v.SwitchToPage("main")
	v.app.SetFocus(v.CurrentPage().Item)
}

func (v *helpView) getTitle() string {
	return v.title
}

func (v *helpView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	switch evt.Key() {
	case tcell.KeyEscape:
		v.back(evt)
		return nil
	case tcell.KeyEnter:
		t := v.GetPrimitive("main").(*tview.Table)
		r, _ := t.GetSelection()
		if r > 0 {
			v.app.command.run(strings.TrimSpace(t.GetCell(r, 0).Text))
			return nil
		}
	}
	return evt
}

func (v *helpView) back(evt *tcell.EventKey) {
	v.app.inject(v.current)
}

func (v *helpView) hints() hints {
	return v.keys.toHints()
}

package view

import (
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type portSelector struct {
	title, port string
	ok, cancel  func()
}

func (p *portSelector) show(app *App) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	f1 := p.port
	f.AddInputField("Pod Port:", f1, 20, nil, func(changed string) {
		f1 = changed
	})

	f.AddButton("OK", p.ok)
	f.AddButton("Cancel", p.cancel)

	modal := tview.NewModalForm("<"+p.title+">", f)
	modal.SetDoneFunc(func(_ int, b string) {
		p.cancel()
	})
}

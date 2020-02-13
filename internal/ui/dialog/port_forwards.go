package dialog

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

type PortForwardFunc func(path, address, lport, cport string)

// ShowPortForwards pops a port forwarding configuration dialog.
func ShowPortForwards(p *ui.Pages, s *config.Styles, path string, ports []string, okFn PortForwardFunc) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(s.BgColor()).
		SetButtonTextColor(s.FgColor()).
		SetLabelColor(config.AsColor(s.K9s.Info.FgColor)).
		SetFieldTextColor(config.AsColor(s.K9s.Info.SectionColor))

	p1, p2, address := ports[0], ports[0], "localhost"
	f.AddDropDown("Container Ports", ports, 0, func(sel string, _ int) {
		p1, p2 = sel, stripPort(sel)
	})

	dropD, ok := f.GetFormItem(0).(*tview.DropDown)
	if ok {
		dropD.SetFieldBackgroundColor(s.BgColor())
		list := dropD.GetList()
		list.SetMainTextColor(s.FgColor())
		list.SetSelectedTextColor(s.FgColor())
		list.SetSelectedBackgroundColor(config.AsColor(s.Table().CursorColor))
		list.SetBackgroundColor(s.BgColor() + 100)
	}
	f.AddInputField("Local Port:", p2, 20, nil, func(p string) {
		p2 = p
	})
	f.AddInputField("Address:", address, 20, nil, func(h string) {
		address = h
	})

	f.AddButton("OK", func() {
		okFn(path, address, stripPort(p2), stripPort(p1))
	})
	f.AddButton("Cancel", func() {
		DismissPortForward(p)
	})

	modal := tview.NewModalForm(fmt.Sprintf("<PortForward on %s>", path), f)
	modal.SetDoneFunc(func(_ int, b string) {
		DismissPortForward(p)
	})
	p.AddPage(portForwardKey, modal, false, false)
	p.ShowPage(portForwardKey)
}

// DismissPortForward dismiss the port forward dialog.
func DismissPortForwards(p *ui.Pages) {
	p.RemovePage(portForwardKey)
}

package dialog

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const portForwardKey = "portforward"

// PortForwardFunc represents a port-forward callback function.
type PortForwardFunc func(path, co string, mapper dao.Tunnel)

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
		p1, p2 = sel, extractPort(sel)
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
		tunnel := dao.Tunnel{
			Address:       address,
			LocalPort:     p2,
			ContainerPort: extractPort(p1),
		}
		okFn(path, extractContainer(p1), tunnel)
	})
	f.AddButton("Cancel", func() {
		DismissPortForwards(p)
	})

	modal := tview.NewModalForm(fmt.Sprintf("<PortForward on %s>", path), f)
	modal.SetDoneFunc(func(_ int, b string) {
		DismissPortForwards(p)
	})
	p.AddPage(portForwardKey, modal, false, false)
	p.ShowPage(portForwardKey)
}

// DismissPortForwards dismiss the port forward dialog.
func DismissPortForwards(p *ui.Pages) {
	p.RemovePage(portForwardKey)
}

// ----------------------------------------------------------------------------
// Helpers...

func extractPort(p string) string {
	tokens := strings.Split(p, ":")
	switch {
	case len(tokens) < 2:
		return tokens[0]
	case len(tokens) == 2:
		return strings.Replace(tokens[1], "╱UDP", "", 1)
	default:
		return tokens[1]
	}
}

func extractContainer(p string) string {
	tokens := strings.Split(p, ":")
	if len(tokens) != 2 {
		return "n/a"
	}

	co, _ := client.Namespaced(tokens[0])
	return co
}

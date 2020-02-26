package view

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const portForwardKey = "portforward"

// PortForwardFunc represents a port-forward callback function.
type PortForwardFunc func(v ResourceViewer, path, co string, mapper client.PortTunnel)

// ShowPortForwards pops a port forwarding configuration dialog.
func ShowPortForwards(v ResourceViewer, path string, ports []string, okFn PortForwardFunc) {
	styles := v.App().Styles

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.BgColor()).
		SetButtonTextColor(styles.FgColor()).
		SetLabelColor(styles.K9s.Info.FgColor.Color()).
		SetFieldTextColor(styles.K9s.Info.SectionColor.Color())

	p1, p2, address := ports[0], extractPort(ports[0]), "localhost"
	f.AddInputField("Container Port:", p1, 30, nil, func(p string) {
		p1 = p
	})
	f.AddInputField("Local Port:", p2, 30, nil, func(p string) {
		p2 = p
	})
	f.AddInputField("Address:", address, 30, nil, func(h string) {
		address = h
	})

	pages := v.App().Content.Pages

	f.AddButton("OK", func() {
		tunnel := client.PortTunnel{
			Address:       address,
			LocalPort:     p2,
			ContainerPort: extractPort(p1),
		}
		okFn(v, path, extractContainer(p1), tunnel)
	})
	f.AddButton("Cancel", func() {
		DismissPortForwards(v, pages)
	})

	modal := tview.NewModalForm(fmt.Sprintf("<PortForward on %s>", path), f)
	modal.SetText("Exposed Ports: " + strings.Join(ports, ","))
	modal.SetDoneFunc(func(_ int, b string) {
		DismissPortForwards(v, pages)
	})

	pages.AddPage(portForwardKey, modal, false, true)
	pages.ShowPage(portForwardKey)
	v.App().SetFocus(pages.GetPrimitive(portForwardKey))
}

// DismissPortForwards dismiss the port forward dialog.
func DismissPortForwards(v ResourceViewer, p *ui.Pages) {
	p.RemovePage(portForwardKey)
	v.App().SetFocus(p.CurrentPage().Item)
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

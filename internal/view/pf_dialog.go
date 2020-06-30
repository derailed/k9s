package view

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const portForwardKey = "portforward"

// PortForwardCB represents a port-forward callback function.
type PortForwardCB func(v ResourceViewer, path, co string, mapper []client.PortTunnel)

// ShowPortForwards pops a port forwarding configuration dialog.
func ShowPortForwards(v ResourceViewer, path string, ports []string, okFn PortForwardCB) {
	styles := v.App().Styles

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.BgColor()).
		SetButtonTextColor(styles.FgColor()).
		SetLabelColor(styles.K9s.Info.FgColor.Color()).
		SetFieldTextColor(styles.K9s.Info.SectionColor.Color())

	address := v.App().Config.CurrentCluster().PortForwardAddress
	p1, p2 := ports[0], extractPort(ports[0])
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
		pp1 := strings.Split(p1, ",")
		pp2 := strings.Split(p2, ",")
		if len(pp1) == 0 || len(pp1) != len(pp2) {
			v.App().Flash().Err(fmt.Errorf("container to local port mismatch"))
			return
		}

		for _, p := range pp1 {
			if !hasPort(p, ports) {
				v.App().Flash().Err(fmt.Errorf("container port must match exposed ports"))
				return
			}
		}

		var tt []client.PortTunnel
		for i := range pp1 {
			tt = append(tt, client.PortTunnel{
				Address:       address,
				LocalPort:     pp2[i],
				ContainerPort: extractPort(pp1[i]),
			})
		}
		okFn(v, path, extractContainer(pp1[0]), tt)
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

func hasPort(port string, pp []string) bool {
	for _, p := range pp {
		if p != port {
			return false
		}
	}

	return true
}

func extractPort(p string) string {
	rx := regexp.MustCompile(`\A([\w|-]+)/?([\w|-]+)?:?(\d+)?(╱UDP)?\z`)
	mm := rx.FindStringSubmatch(p)
	if len(mm) != 5 {
		return p
	}
	for i := 3; i > 0; i-- {
		if mm[i] != "" {
			return mm[i]
		}
	}
	return p
}

func extractContainer(p string) string {
	tokens := strings.Split(p, ":")
	if len(tokens) != 2 {
		return "n/a"
	}

	co, _ := client.Namespaced(tokens[0])
	return co
}

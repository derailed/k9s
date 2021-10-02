package view

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const portForwardKey = "portforward"

// PortForwardCB represents a port-forward callback function.
type PortForwardCB func(v ResourceViewer, path, co string, mapper []client.PortTunnel)

// ShowPortForwards pops a port forwarding configuration dialog.
func ShowPortForwards(v ResourceViewer, path string, ports []string, ann string, okFn PortForwardCB) {
	styles := v.App().Styles.Dialog()

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())

	address := v.App().Config.CurrentCluster().PortForwardAddress

	var p1, p2 string
	if len(ports) > 0 {
		p1, p2 = ports[0], extractPort(ports[0])
		if len(ann) != 0 {
			container, port, ok := parsePFAnn(ann)
			if ok {
				for _, p := range ports {
					co, po, portNum := parsePort(p)
					if co == container && port == po || port == portNum {
						p1, p2 = p, extractPort(p)
						break
					}
				}
			}
		}
	}
	fieldLen := int(math.Max(30, float64(len(p1))))
	f.AddInputField("Container Port:", p1, fieldLen, nil, func(p string) {
		p1 = p
	})
	field := f.GetFormItemByLabel("Container Port:").(*tview.InputField)
	if field.GetText() == "" {
		field.SetPlaceholder("Enter a container name/port")
	}
	f.AddInputField("Local Port:", p2, fieldLen, nil, func(p string) {
		p2 = p
	})
	field = f.GetFormItemByLabel("Local Port:").(*tview.InputField)
	if field.GetText() == "" {
		field.SetPlaceholder("Enter a local port")
	}
	f.AddInputField("Address:", address, fieldLen, nil, func(h string) {
		address = h
	})
	for i := 0; i < 3; i++ {
		field, ok := f.GetFormItem(i).(*tview.InputField)
		if !ok {
			continue
		}
		field.SetLabelColor(styles.LabelFgColor.Color())
		field.SetFieldTextColor(styles.FieldFgColor.Color())
	}

	f.AddButton("OK", func() {
		pp1 := strings.Split(p1, ",")
		pp2 := strings.Split(p2, ",")
		if len(pp1) == 0 || len(pp1) != len(pp2) {
			v.App().Flash().Err(fmt.Errorf("container to local port mismatch"))
			return
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
	pages := v.App().Content.Pages
	f.AddButton("Cancel", func() {
		DismissPortForwards(v, pages)
	})
	for i := 0; i < 2; i++ {
		b := f.GetButton(i)
		if b == nil {
			continue
		}
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	modal := tview.NewModalForm("<PortForward>", f)
	msg := path
	if len(ports) > 1 {
		msg += "\n\nExposed Ports:\n" + strings.Join(ports, "\n")
	}
	modal.SetText(msg)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetBackgroundColor(styles.BgColor.Color())
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

func parsePort(p string) (string, string, string) {
	rx := regexp.MustCompile(`\A([\w|-]+)/?([\w|-]+)?:?(\d+)?(╱UDP)?\z`)
	mm := rx.FindStringSubmatch(p)
	if len(mm) != 5 {
		return "", "", ""
	}

	return mm[1], mm[2], mm[3]
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
		return render.NAValue
	}

	co, _ := client.Namespaced(tokens[0])
	return co
}

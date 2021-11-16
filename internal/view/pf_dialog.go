package view

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/port"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

const portForwardKey = "portforward"

// PortForwardCB represents a port-forward callback function.
type PortForwardCB func(ResourceViewer, string, port.PortTunnels) error

// ShowPortForwards pops a port forwarding configuration dialog.
func ShowPortForwards(v ResourceViewer, path string, ports port.ContainerPortSpecs, aa port.Annotations, okFn PortForwardCB) {
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

	pf, err := aa.PreferredPorts(ports)
	if err != nil {
		log.Warn().Err(err).Msgf("unable to resolve ports")
	}

	p1, p2 := pf.ToPortSpec(ports)
	fieldLen := int(math.Max(30, float64(len(p1))))
	f.AddInputField("Container Port:", p1, fieldLen, nil, nil)
	coField := f.GetFormItemByLabel("Container Port:").(*tview.InputField)
	if coField.GetText() == "" {
		coField.SetPlaceholder("Enter a container name/port")
	}
	f.AddInputField("Local Port:", p2, fieldLen, nil, nil)
	poField := f.GetFormItemByLabel("Local Port:").(*tview.InputField)
	if poField.GetText() == "" {
		poField.SetPlaceholder("Enter a local port")
	}
	coField.SetChangedFunc(func(s string) {
		port := extractPort(s)
		poField.SetText(port)
		p2 = port
	})
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
		if coField.GetText() == "" || poField.GetText() == "" {
			v.App().Flash().Err(fmt.Errorf("container to local port mismatch"))
			return
		}
		tt, err := port.ToTunnels(address, coField.GetText(), poField.GetText())
		if err != nil {
			v.App().Flash().Err(err)
			return
		}
		if err := okFn(v, path, tt); err != nil {
			v.App().Flash().Err(err)
		}
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
		msg += "\n\nExposed Ports:\n" + ports.Dump()
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

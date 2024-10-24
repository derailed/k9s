// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/port"
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

	ct, err := v.App().Config.K9s.ActiveContext()
	if err != nil {
		log.Error().Err(err).Msgf("No active context detected")
		return
	}

	pf, err := aa.PreferredPorts(ports)
	if err != nil {
		log.Warn().Err(err).Msgf("unable to resolve ports on %s", path)
	}

	p1, p2 := pf.ToPortSpec(ports)
	fieldLen := int(math.Max(30, float64(len(p1))))
	f.AddInputField("Container Port:", p1, fieldLen, nil, nil)
	f.AddInputField("Local Port:", p2, fieldLen, nil, nil)
	coField := f.GetFormItemByLabel("Container Port:").(*tview.InputField)
	loField := f.GetFormItemByLabel("Local Port:").(*tview.InputField)
	if coField.GetText() == "" {
		coField.SetPlaceholder("Enter a container name::port")
	}
	coField.SetChangedFunc(func(s string) {
		port := extractPort(s)
		loField.SetText(port)
		p2 = port
	})
	if loField.GetText() == "" {
		loField.SetPlaceholder("Enter a local port")
	}
	address := ct.PortForwardAddress
	f.AddInputField("Address:", address, fieldLen, nil, func(h string) {
		address = h
	})
	for i := 0; i < 3; i++ {
		if field, ok := f.GetFormItem(i).(*tview.InputField); ok {
			field.SetLabelColor(styles.LabelFgColor.Color())
			field.SetFieldTextColor(styles.FieldFgColor.Color())
		}
	}

	f.AddButton("OK", func() {
		if coField.GetText() == "" || loField.GetText() == "" {
			v.App().Flash().Err(fmt.Errorf("container to local port mismatch"))
			return
		}
		tt, err := port.ToTunnels(address, coField.GetText(), loField.GetText())
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
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
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

func extractPort(port string) string {
	tokens := strings.Split(port, "::")
	if len(tokens) < 2 {
		ports := strings.Split(port, ",")
		for _, t := range ports {
			if _, err := strconv.Atoi(strings.TrimSpace(t)); err != nil {
				return ""
			}
		}
		return port
	}

	return tokens[1]
}

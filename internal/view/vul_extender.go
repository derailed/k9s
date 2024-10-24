// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// VulnerabilityExtender adds vul image scan extensions.
type VulnerabilityExtender struct {
	ResourceViewer
}

// NewVulnerabilityExtender returns a new extender.
func NewVulnerabilityExtender(r ResourceViewer) ResourceViewer {
	v := VulnerabilityExtender{ResourceViewer: r}
	v.AddBindKeysFn(v.bindKeys)

	return &v
}

func (v *VulnerabilityExtender) bindKeys(aa *ui.KeyActions) {
	if v.App().Config.K9s.ImageScans.Enable {
		aa.Bulk(ui.KeyMap{
			ui.KeyV:      ui.NewKeyAction("Show Vulnerabilities", v.showVulCmd, true),
			ui.KeyShiftV: ui.NewKeyAction("Sort Vulnerabilities", v.GetTable().SortColCmd("VS", true), false),
		})
	}
}

func (v *VulnerabilityExtender) showVulCmd(evt *tcell.EventKey) *tcell.EventKey {
	isv := NewImageScan(client.NewGVR("scans"))
	isv.SetContextFn(v.selContext)
	if err := v.App().inject(isv, false); err != nil {
		v.App().Flash().Err(err)
	}

	return nil
}

func (v *VulnerabilityExtender) selContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyPath, v.GetTable().GetSelectedItem())
	return context.WithValue(ctx, internal.KeyGVR, v.GVR())
}

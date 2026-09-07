// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
)

// Probe represents a Prometheus Operator Probe viewer.
type Probe struct {
	ResourceViewer
}

// NewProbe returns a new Probe viewer.
func NewProbe(gvr *client.GVR) ResourceViewer {
	p := Probe{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	p.GetTable().SetEnterFn(p.showIngresses)

	return &p
}

func (p *Probe) showIngresses(a *App, _ ui.Tabular, _ *client.GVR, path string) {
	var d dao.Probe
	d.Init(a.factory, p.GVR())

	tgt, err := d.SelectedTarget(path)
	if err != nil {
		a.Flash().Err(err)
		return
	}
	if tgt == nil {
		a.Flash().Warnf("Probe %s has no ingress selector (static targets only)", path)
		return
	}
	if tgt.Selector.Empty() {
		a.Flash().Warnf("Probe %s has an empty ingress selector", path)
		return
	}
	showIngresses(a, path, tgt.Namespace, tgt.Selector)
}

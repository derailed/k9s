// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
)

// PodMonitor represents a Prometheus Operator PodMonitor viewer.
type PodMonitor struct {
	ResourceViewer
}

// NewPodMonitor returns a new PodMonitor viewer.
func NewPodMonitor(gvr *client.GVR) ResourceViewer {
	p := PodMonitor{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	p.GetTable().SetEnterFn(p.showPods)

	return &p
}

func (p *PodMonitor) showPods(a *App, _ ui.Tabular, _ *client.GVR, path string) {
	var d dao.PodMonitor
	d.Init(a.factory, p.GVR())

	tgt, err := d.SelectedTarget(path)
	if err != nil {
		a.Flash().Err(err)
		return
	}
	if tgt.Selector.Empty() {
		a.Flash().Warnf("PodMonitor %s has no pod selector", path)
		return
	}
	showPodsIn(a, path, tgt.Namespace, tgt.Selector)
}

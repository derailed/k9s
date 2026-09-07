// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
)

// ServiceMonitor represents a Prometheus Operator ServiceMonitor viewer.
type ServiceMonitor struct {
	ResourceViewer
}

// NewServiceMonitor returns a new ServiceMonitor viewer.
func NewServiceMonitor(gvr *client.GVR) ResourceViewer {
	s := ServiceMonitor{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	s.GetTable().SetEnterFn(s.showServices)

	return &s
}

func (s *ServiceMonitor) showServices(a *App, _ ui.Tabular, _ *client.GVR, path string) {
	var d dao.ServiceMonitor
	d.Init(a.factory, s.GVR())

	tgt, err := d.SelectedTarget(path)
	if err != nil {
		a.Flash().Err(err)
		return
	}
	if tgt.Selector.Empty() {
		a.Flash().Warnf("ServiceMonitor %s has no service selector", path)
		return
	}
	showServices(a, path, tgt.Namespace, tgt.Selector)
}

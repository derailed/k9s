// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
)

// CRD represents a crd viewer.
type CRD struct {
	ResourceViewer
}

// NewCRD returns a new viewer.
func NewCRD(gvr client.GVR) ResourceViewer {
	s := CRD{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	s.AddBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showCRD)

	return &s
}

func (s *CRD) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftV: ui.NewKeyAction("Sort Versions", s.GetTable().SortColCmd("VERSIONS", false), true),
		ui.KeyShiftR: ui.NewKeyAction("Sort Group", s.GetTable().SortColCmd("GROUP", true), true),
		ui.KeyShiftK: ui.NewKeyAction("Sort Kind", s.GetTable().SortColCmd("KIND", true), true),
	})
}

func (s *CRD) showCRD(app *App, _ ui.Tabular, _ client.GVR, path string) {
	_, crd := client.Namespaced(path)
	app.gotoResource(crd, "", false)
}

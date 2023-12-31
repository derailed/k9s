// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// PersistentVolumeClaim represents a PVC custom viewer.
type PersistentVolumeClaim struct {
	ResourceViewer
}

// NewPersistentVolumeClaim returns a new viewer.
func NewPersistentVolumeClaim(gvr client.GVR) ResourceViewer {
	v := PersistentVolumeClaim{
		ResourceViewer: NewBrowser(gvr),
	}
	v.AddBindKeysFn(v.bindKeys)

	return &v
}

func (p *PersistentVolumeClaim) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyU:      ui.NewKeyAction("UsedBy", p.refCmd, true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", p.GetTable().SortColCmd("STATUS", true), false),
		ui.KeyShiftV: ui.NewKeyAction("Sort Volume", p.GetTable().SortColCmd("VOLUME", true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort StorageClass", p.GetTable().SortColCmd("STORAGECLASS", true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Capacity", p.GetTable().SortColCmd("CAPACITY", true), false),
	})
}

func (p *PersistentVolumeClaim) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, p.App(), p.GetTable(), dao.PvcGVR)
}

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
)

// PersistentVolumeClaim represents a PVC custom viewer.
type PersistentVolumeClaim struct {
	ResourceViewer
}

// NewPersistentVolumeClaim returns a new viewer.
func NewPersistentVolumeClaim(gvr client.GVR) ResourceViewer {
	d := PersistentVolumeClaim{
		ResourceViewer: NewBrowser(gvr),
	}
	d.SetBindKeysFn(d.bindKeys)
	d.GetTable().SetColorerFn(render.PersistentVolumeClaim{}.ColorerFunc())

	return &d
}

func (d *PersistentVolumeClaim) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", d.GetTable().SortColCmd("STATUS", true), false),
		ui.KeyShiftV: ui.NewKeyAction("Sort Volume", d.GetTable().SortColCmd("VOLUME", true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort StorageClass", d.GetTable().SortColCmd("STORAGECLASS", true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Capacity", d.GetTable().SortColCmd("CAPACITY", true), false),
	})
}

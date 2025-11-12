// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// Event represents a command alias view.
type Event struct {
	ResourceViewer
}

// NewEvent returns a new alias view.
func NewEvent(gvr *client.GVR) ResourceViewer {
	e := Event{
		ResourceViewer: NewBrowser(gvr),
	}
	e.AddBindKeysFn(e.bindKeys)
	e.GetTable().SetSortCol("LAST SEEN", false)

	return &e
}

func (*Event) bindKeys(aa *ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlD, ui.KeyE, ui.KeyA)
}

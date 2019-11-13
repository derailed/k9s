package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// ScalableResource represents a resource that can be scaled.
type ScalableResource struct {
	*Resource
}

// NewScalableResource returns a new viewer.
func NewScalableResource(title, gvr string, list resource.List) ResourceViewer {
	return newScalableResourceForParent(NewResource(title, gvr, list))
}

func newScalableResourceForParent(parent *Resource) *ScalableResource {
	s := ScalableResource{
		Resource: parent,
	}
	parent.extraActionsFn = s.extraActions

	return &s
}

func (s *ScalableResource) extraActions(aa ui.KeyActions) {
	aa[ui.KeyS] = ui.NewKeyAction("Scale", s.scaleCmd, true)
}

func (s *ScalableResource) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.masterPage().RowSelected() {
		return evt
	}

	s.showScaleDialog(s.list.GetName(), s.masterPage().GetSelectedItem())
	return nil
}

func (s *ScalableResource) scale(selection string, replicas int) {
	ns, n := namespaced(selection)

	r := s.list.Resource().(resource.Scalable)

	err := r.Scale(ns, n, int32(replicas))
	if err != nil {
		s.app.Flash().Err(err)
	}
}

func (s *ScalableResource) showScaleDialog(resourceType string, resourceName string) {
	f := s.createScaleForm()

	confirm := tview.NewModalForm("<Scale>", f)
	confirm.SetText(fmt.Sprintf("Scale %s %s", resourceType, resourceName))
	confirm.SetDoneFunc(func(int, string) {
		s.dismissScaleDialog()
	})
	s.AddPage(scaleDialogKey, confirm, false, false)
	s.ShowPage(scaleDialogKey)
}

func (s *ScalableResource) createScaleForm() *tview.Form {
	f := s.createStyledForm()

	tv := s.masterPage()
	replicas := strings.TrimSpace(tv.GetCell(tv.GetSelectedRowIndex(), tv.NameColIndex()+1).Text)
	f.AddInputField("Replicas:", replicas, 4, func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		replicas = changed
	})

	f.AddButton("OK", func() {
		s.okSelected(replicas)
	})

	f.AddButton("Cancel", func() {
		s.dismissScaleDialog()
	})

	return f
}

func (s *ScalableResource) createStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)
	return f
}

func (s *ScalableResource) okSelected(replicas string) {
	if val, err := strconv.Atoi(replicas); err == nil {
		s.scale(s.masterPage().GetSelectedItem(), val)
	} else {
		s.app.Flash().Err(err)
	}

	s.dismissScaleDialog()
}

func (s *ScalableResource) dismissScaleDialog() {
	s.Pages.RemovePage(scaleDialogKey)
}

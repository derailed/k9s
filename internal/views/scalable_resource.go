package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type (
	scalableResourceView struct {
		*resourceView
	}
)

func newScalableResourceView(title string, app *appView, list resource.List) resourceViewer {
	return *newScalableResourceViewForParent(newResourceView(title, app, list).(*resourceView))
}

func newScalableResourceViewForParent(parent *resourceView) *scalableResourceView {
	v := scalableResourceView{
		parent,
	}
	parent.extraActionsFn = v.extraActions
	return &v
}

func (v *scalableResourceView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyS] = ui.NewKeyAction("Scale", v.scaleCmd, true)
}

func (v *scalableResourceView) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	v.showScaleDialog(v.list.GetName(), v.masterPage().GetSelectedItem())
	return nil
}

func (v *scalableResourceView) scale(selection string, replicas int) {
	ns, n := namespaced(selection)

	r := v.list.Resource().(resource.Scalable)

	err := r.Scale(ns, n, int32(replicas))
	if err != nil {
		v.app.Flash().Err(err)
	}
}

func (v *scalableResourceView) showScaleDialog(resourceType string, resourceName string) {
	f := v.createScaleForm()

	confirm := tview.NewModalForm("<Scale>", f)
	confirm.SetText(fmt.Sprintf("Scale %s %s", resourceType, resourceName))
	confirm.SetDoneFunc(func(int, string) {
		v.dismissScaleDialog()
	})
	v.AddPage(scaleDialogKey, confirm, false, false)
	v.ShowPage(scaleDialogKey)
}

func (v *scalableResourceView) createScaleForm() *tview.Form {
	f := v.createStyledForm()

	tv := v.masterPage()
	replicas := strings.TrimSpace(tv.GetCell(tv.GetSelectedRow(), tv.NameColIndex()+1).Text)
	f.AddInputField("Replicas:", replicas, 4, func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		replicas = changed
	})

	f.AddButton("OK", func() {
		v.okSelected(replicas)
	})

	f.AddButton("Cancel", func() {
		v.dismissScaleDialog()
	})

	return f
}

func (v *scalableResourceView) createStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)
	return f
}

func (v *scalableResourceView) okSelected(replicas string) {
	if val, err := strconv.Atoi(replicas); err == nil {
		v.scale(v.masterPage().GetSelectedItem(), val)
	} else {
		v.app.Flash().Err(err)
	}

	v.dismissScaleDialog()
}

func (v *scalableResourceView) dismissScaleDialog() {
	v.Pages.RemovePage(scaleDialogKey)
}

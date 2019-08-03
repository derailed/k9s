package views

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployView struct {
	*logResourceView
	scalableResourceView *scalableResourceView
}

const scaleDialogKey = "scale"

func newDeployView(title string, app *appView, list resource.List) resourceViewer {
	logResourceView := newLogResourceView(title, app, list)
	v := deployView{logResourceView, newScalableResourceViewForParent(logResourceView.resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *deployView) extraActions(aa keyActions) {
	v.logResourceView.extraActions(aa)
	v.scalableResourceView.extraActions(aa)
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
}

func (v *deployView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDeployment(app.conn())
	dep, err := d.Get(ns, n)
	if err != nil {
		app.flash().err(err)
		return
	}

	dp := dep.(*v1.Deployment)
	l, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		app.flash().err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}

func (v *deployView) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	v.showScaleDialog(v.getList().GetName(), v.getSelectedItem())
	return nil
}

func (v *deployView) scale(selection string, replicas int) {
	ns, n := namespaced(selection)
	d := k8s.NewDeployment(v.app.conn())

	err := d.Scale(ns, n, int32(replicas))
	if err != nil {
		v.app.flash().err(err)
	}
}

func (v *deployView) showScaleDialog(resourceType string, resourceName string) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	replicas := "1"
	f.AddInputField("Replicas:", replicas, 4, func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		replicas = changed
	})

	dismiss := func() {
		v.Pages.RemovePage(scaleDialogKey)
	}

	f.AddButton("OK", func() {
		if val, err := strconv.Atoi(replicas); err == nil {
			v.scale(v.getSelectedItem(), val)
		} else {
			v.app.flash().err(err)
		}

		dismiss()
	})

	f.AddButton("Cancel", func() {
		dismiss()
	})

	confirm := tview.NewModalForm("<Scale>", f)
	confirm.SetText(fmt.Sprintf("Scale %s %s", resourceType, resourceName))
	confirm.SetDoneFunc(func(int, string) {
		dismiss()
	})
	v.AddPage(scaleDialogKey, confirm, false, false)
	v.ShowPage(scaleDialogKey)
}

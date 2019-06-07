package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type jobView struct {
	*logResourceView
}

func newJobView(t string, app *appView, list resource.List) resourceViewer {
	v := jobView{newLogResourceView(t, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

// Handlers...

// func (v *jobView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if v.viewLogs(false) {
// 		return nil
// 	}
// 	return evt
// }

// func (v *jobView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if v.viewLogs(true) {
// 		return nil
// 	}
// 	return evt
// }

// func (v *jobView) viewLogs(previous bool) bool {
// 	if !v.rowSelected() {
// 		return false
// 	}

// 	cc, err := fetchContainers(v.list, v.selectedItem, true)
// 	if err != nil {
// 		v.app.flash().err(err)
// 		log.Error().Err(err).Msgf("Unable to fetch containers for %s", v.selectedItem)
// 		return false
// 	}

// 	if len(cc) == 1 {
// 		v.showLogs(v.selectedItem, cc[0], v.list.GetName(), v, previous)
// 		return true
// 	}

// 	picker := v.GetPrimitive("picker").(*selectList)
// 	picker.populate(cc)
// 	picker.SetSelectedFunc(func(i int, t, d string, r rune) {
// 		v.showLogs(v.selectedItem, t, "picker", picker, previous)
// 	})
// 	v.switchPage("picker")

// 	return true
// }

// func (v *jobView) showLogs(path, co, view string, parent loggable, prev bool) {
// 	l := v.GetPrimitive("logs").(*logsView)
// 	l.reload(co, parent, view, prev)
// 	v.switchPage("logs")
// }

func (v *jobView) extraActions(aa keyActions) {
	v.logResourceView.extraActions(aa)
	// aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	// aa[KeyShiftL] = newKeyAction("Logs Previous", v.prevLogsCmd, true)
}

func (v *jobView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	j := k8s.NewJob(app.conn())
	job, err := j.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching Job %s", sel)
		app.flash().err(err)
		return
	}

	jo := job.(*batchv1.Job)
	l, err := metav1.LabelSelectorAsSelector(jo.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for Job %s", sel)
		app.flash().err(err)
		return
	}

	showPods(app, "", "Job", sel, l.String(), "", v.backCmd)
}

// func (v *jobView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	v.app.inject(v)

// 	return nil
// }

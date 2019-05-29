package views

import (
	"context"
	"errors"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

type containerView struct {
	*resourceView

	current igniter
	exitFn  func()
}

func newContainerView(t string, app *appView, list resource.List, path string, exitFn func()) resourceViewer {
	v := containerView{resourceView: newResourceView(t, app, list).(*resourceView)}
	{
		v.path = &path
		v.extraActionsFn = v.extraActions
		v.colorerFn = containerColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.exitFn = exitFn
	}
	v.AddPage("logs", newLogsView(list.GetName(), &v), true, false)

	return &v
}

func (v *containerView) init(ctx context.Context, ns string) {
	v.resourceView.init(ctx, ns)
	// v.selChanged(1, 0)
}

func (v *containerView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftF] = newKeyAction("PortForward", v.portFwdCmd, true)
	aa[KeyShiftL] = newKeyAction("Logs Previous", v.prevLogsCmd, true)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd, true)
	aa[tcell.KeyEscape] = newKeyAction("Back", v.backCmd, false)
	aa[KeyP] = newKeyAction("Previous", v.backCmd, false)
	aa[tcell.KeyEnter] = newKeyAction("View Logs", v.logsCmd, false)
	aa[KeyShiftC] = newKeyAction("Sort CPU", v.sortColCmd(6, false), true)
	aa[KeyShiftM] = newKeyAction("Sort MEM", v.sortColCmd(7, false), true)
	aa[KeyAltC] = newKeyAction("Sort CPU%", v.sortColCmd(8, false), true)
	aa[KeyAltM] = newKeyAction("Sort MEM%", v.sortColCmd(9, false), true)
}

// Protocol...

func (v *containerView) backFn() actionHandler {
	return v.backCmd
}

func (v *containerView) appView() *appView {
	return v.app
}

func (v *containerView) getList() resource.List {
	return v.list
}

func (v *containerView) getSelection() string {
	return *v.path
}

// Handlers...

func (v *containerView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	cell := v.getTV().GetCell(v.selectedRow, 3)
	if cell != nil && strings.TrimSpace(cell.Text) != "Running" {
		v.app.flash().err(errors.New("No logs for a non running container"))
		return evt
	}
	v.showLogs(v.selectedItem, v.list.GetName(), v, false)

	return nil
}

func (v *containerView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.showLogs(v.selectedItem, v.list.GetName(), v, true)

	return nil
}

func (v *containerView) showLogs(co, view string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, view, prev)
	v.switchPage("logs")
}

func (v *containerView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	v.stopUpdates()
	// v.suspend()
	// {
	shellIn(v.app, *v.path, v.selectedItem)
	// }
	// v.resume()
	v.restartUpdates()
	return nil
}

func (v *containerView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *containerView) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	portC := v.getTV().GetCell(v.selectedRow, 10)
	ports := strings.Split(portC.Text, ",")
	if len(ports) == 0 {
		v.app.flash().err(errors.New("Container exposes no ports"))
		return nil
	}
	port := strings.TrimSpace(ports[0])
	if port == "" {
		v.app.flash().err(errors.New("Container exposed no ports"))
		return nil
	}
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	f1, f2 := port, port
	f.AddInputField("Pod Port:", f1, 10, nil, func(changed string) {
		f1 = changed
	})
	f.AddInputField("Local Port:", f2, 10, nil, func(changed string) {
		f2 = changed
	})

	f.AddButton("OK", func() {
		pf := k8s.NewPortForward(v.app.conn(), &log.Logger)
		ports := []string{f2 + ":" + f1}
		co := strings.TrimSpace(v.getTV().GetCell(v.selectedRow, 0).Text)
		fw, err := pf.Start(*v.path, co, ports)
		if err != nil {
			log.Error().Err(err).Msg("Fort Forward")
			v.app.flash().errf("PortForward failed! %v", err)
			return
		}

		log.Debug().Msgf(">>> Starting port forward %q %v", *v.path, ports)
		go func(f *portforward.PortForwarder) {
			v.app.QueueUpdate(func() {
				v.app.forwarders = append(v.app.forwarders, pf)
				v.app.flash().infof("PortForward activated %s:%s", pf.Path(), pf.Ports()[0])
				v.app.gotoResource("pf", true)
			})
			pf.SetActive(true)
			if err := f.ForwardPorts(); err != nil {
				v.app.QueueUpdate(func() {
					if len(v.app.forwarders) > 0 {
						v.app.forwarders = v.app.forwarders[:len(v.app.forwarders)-1]
					}
					pf.SetActive(false)
					log.Error().Err(err).Msg("Port forward failed")
					v.app.flash().errf("PortForward failed %s", err)
				})
			}
		}(fw)
	})
	f.AddButton("Cancel", func() {
		v.app.flash().info("Canceled!!")
		v.dismissModal()
	})

	modal := tview.NewModalForm("<PortForward>", f)
	modal.SetDoneFunc(func(_ int, b string) {
		v.dismissModal()
	})
	v.AddPage("dialog", modal, false, false)
	v.ShowPage("dialog")

	return nil
}

func (v *containerView) dismissModal() {
	v.RemovePage("dialog")
	v.switchPage(v.list.GetName())
}

func (v *containerView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.exitFn()

	return nil
}

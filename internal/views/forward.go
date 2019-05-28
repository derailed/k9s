package views

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	forwardTitle    = "Port Forwards"
	forwardTitleFmt = " [aqua::b]%s([fuchsia::b]%d[fuchsia::-])[aqua::-] "
)

type forwardView struct {
	*tview.Pages

	app     *appView
	current igniter
	cancel  context.CancelFunc
	bench   *benchmark
}

func newForwardView(app *appView) *forwardView {
	v := forwardView{
		Pages: tview.NewPages(),
		app:   app,
	}

	tv := newTableView(app, forwardTitle)
	tv.SetBorderFocusColor(tcell.ColorDodgerBlue)
	tv.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDodgerBlue, tcell.AttrNone)
	tv.colorerFn = forwardColorer
	tv.currentNS = ""
	v.AddPage("table", tv, true, true)

	v.current = app.content.GetPrimitive("main").(igniter)
	v.registerActions()

	return &v
}

// Init the view.
func (v *forwardView) init(context.Context, string) {
	tv := v.getTV()
	v.refresh()
	tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+4, true
	tv.refresh()
	tv.Select(1, 0)
	v.app.SetFocus(tv)
}

func (v *forwardView) getTV() *tableView {
	if vu, ok := v.GetPrimitive("table").(*tableView); ok {
		return vu
	}

	return nil
}

func (v *forwardView) refresh() {
	tv := v.getTV()
	tv.update(v.hydrate())
	v.app.SetFocus(tv)
	tv.resetTitle()
}

func (v *forwardView) registerActions() {
	tv := v.getTV()
	tv.actions[tcell.KeyCtrlB] = newKeyAction("Bench", v.benchCmd, true)
	tv.actions[KeyAltB] = newKeyAction("Bench Stop", v.benchStopCmd, true)
	tv.actions[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, true)
	tv.actions[KeySlash] = newKeyAction("Filter", tv.activateCmd, false)
	tv.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)
	tv.actions[KeyShiftP] = newKeyAction("Sort Ports", v.sortColCmd(2, true), true)
	tv.actions[KeyShiftU] = newKeyAction("Sort URL", v.sortColCmd(4, true), true)
}

func (v *forwardView) getTitle() string {
	return forwardTitle
}

func (v *forwardView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		tv := v.getTV()
		tv.sortCol.index, tv.sortCol.asc = tv.nameColIndex()+col, asc
		v.refresh()

		return nil
	}
}

func (v *forwardView) benchStopCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.bench != nil {
		log.Debug().Msg(">>> Benchmark canceled!!")
		v.app.flash().info("Benchmark canceled!")
		v.bench.cancel()
		v.bench = nil
	}
	v.app.statusReset()

	return nil
}

func (v *forwardView) benchCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}

	if v.bench != nil {
		v.app.flash().err(errors.New("Only one benchmark allowed at a time"))
		return nil
	}

	tv := v.getTV()
	r, _ := tv.GetSelection()
	url := strings.TrimSpace(tv.GetCell(r, 4).Text)
	log.Debug().Msgf("Go Routines before %d", runtime.NumGoroutine())
	cfg := benchConfig{
		Method: "GET",
		Path:   sel,
		URL:    url,
		C:      1,
		N:      200,
	}
	var err error
	if v.bench, err = newBenchmark(cfg); err != nil {
		log.Error().Err(err).Msg("Bench failed!")
		v.app.flash().errf("Bench failed %v", err)
		v.app.statusReset()
		return nil
	}

	v.app.status(flashWarn, "Starting Benchmark...")
	log.Debug().Msg("Bench starting...")
	go v.bench.run(func() {
		log.Debug().Msg("Bench Completed!")
		v.app.QueueUpdate(func() {
			v.bench = nil
			v.app.flash().infof("Benchmark for %s is done!", sel)
			v.app.status(flashInfo, "Benchmark Completed!")
			go func() {
				<-time.After(2 * time.Second)
				v.app.QueueUpdate(func() {
					v.app.statusReset()
				})
			}()
		})
	})
	log.Debug().Msgf("Go Routines after %d", runtime.NumGoroutine())

	return nil
}

func (v *forwardView) getSelectedItem() string {
	tv := v.getTV()
	r, _ := tv.GetSelection()
	if r == 0 {
		return ""
	}

	return fqn(strings.TrimSpace(tv.GetCell(r, 0).Text), strings.TrimSpace(tv.GetCell(r, 1).Text))
}

func (v *forwardView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	tv := v.getTV()
	if !tv.cmdBuff.empty() {
		tv.cmdBuff.reset()
		return nil
	}

	sel := v.getSelectedItem()
	if sel == "" {
		return nil
	}

	showModal(v.Pages, fmt.Sprintf("Deleting `%s are you sure?", sel), "table", func() {
		index := -1
		for i, f := range v.app.forwarders {
			if sel == f.Path() {
				index = i
			}
		}
		if index == -1 {
			return
		}
		if index == 0 && len(v.app.forwarders) == 1 {
			v.app.forwarders = []forwarder{}
		} else {
			v.app.forwarders = append(v.app.forwarders[:index], v.app.forwarders[index+1:]...)
		}
		v.getTV().update(v.hydrate())
		v.app.flash().infof("PortForward %s deleted!", sel)
	})

	return nil
}

func (v *forwardView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	tv := v.getTV()
	if tv.cmdBuff.isActive() {
		tv.cmdBuff.reset()
	} else {
		v.app.inject(v.current)
	}

	return nil
}

func (v *forwardView) runCmd(evt *tcell.EventKey) *tcell.EventKey {
	tv := v.getTV()
	r, _ := tv.GetSelection()
	if r > 0 {
		v.app.gotoResource(strings.TrimSpace(tv.GetCell(r, 0).Text), true)
	}

	return nil
}

func (v *forwardView) hints() hints {
	return v.getTV().actions.toHints()
}

func (v *forwardView) hydrate() resource.TableData {
	cmds := helpCmds(v.app.conn())

	data := resource.TableData{
		Header:    resource.Row{"NAMESPACE", "NAME", "PORTS", "ACTIVE", "URL", "AGE"},
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: resource.AllNamespaces,
	}

	for _, f := range v.app.forwarders {
		ports := strings.Split(f.Ports()[0], ":")
		ns, n := namespaced(f.Path())
		fields := resource.Row{
			ns,
			n,
			strings.Join(f.Ports(), ","),
			fmt.Sprintf("%t", f.Active()),
			"http://localhost" + ":" + ports[0],
			f.Age(),
		}
		data.Rows[f.Path()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func (v *forwardView) resetTitle() {
	v.SetTitle(fmt.Sprintf(forwardTitleFmt, forwardTitle, v.getTV().GetRowCount()-1))
}

const genericPrompt = "prompt"

func showModal(pv *tview.Pages, msg, back string, ok func()) {
	m := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(func(_ int, b string) {
			if b == "OK" {
				ok()
			}
			dismissModal(pv, back)
		})
	m.SetTitle("<Confirm>")
	pv.AddPage(genericPrompt, m, false, false)
	pv.ShowPage(genericPrompt)
}

func dismissModal(pv *tview.Pages, page string) {
	pv.RemovePage(genericPrompt)
	pv.SwitchToPage(page)
}

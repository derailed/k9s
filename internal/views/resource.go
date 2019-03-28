package views

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	refreshDelay = 0.1
	noSelection  = ""
)

type (
	details interface {
		tview.Primitive
		setTitle(string)
		clear()
		setActions(keyActions)
		update(resource.Properties)
	}

	resourceView struct {
		*tview.Pages

		app            *appView
		title          string
		selectedItem   string
		selectedRow    int
		namespaces     map[int]string
		selectedNS     string
		update         sync.Mutex
		list           resource.List
		enterFn        enterFn
		extraActionsFn func(keyActions)
		selectedFn     func() string
		decorateFn     decorateFn
		colorerFn      colorerFn
	}
)

func newResourceView(title string, app *appView, list resource.List) resourceViewer {
	v := resourceView{
		app:        app,
		title:      title,
		list:       list,
		selectedNS: list.GetNamespace(),
		Pages:      tview.NewPages(),
	}

	tv := newTableView(app, v.title)
	{
		tv.SetSelectionChangedFunc(v.selChanged)
	}
	v.AddPage(v.list.GetName(), tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)

	confirm := tview.NewModal().
		AddButtons([]string{"OK", "Cancel"}).
		SetTextColor(tcell.ColorFuchsia)
	v.AddPage("confirm", confirm, false, false)

	return &v
}

// Init watches all running pods in given namespace
func (v *resourceView) init(ctx context.Context, ns string) {
	v.selectedItem, v.selectedNS = noSelection, ns

	colorer := defaultColorer
	if v.colorerFn != nil {
		colorer = v.colorerFn
	}
	v.getTV().setColorer(colorer)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s watcher canceled!", v.title)
				return
			case <-time.After(time.Duration(v.app.config.K9s.RefreshRate) * time.Second):
				v.refresh()
			}
		}
	}(ctx)
	v.refresh()
	if tv, ok := v.CurrentPage().Item.(*tableView); ok {
		tv.Select(0, 0)
	}
}

func (v *resourceView) getTitle() string {
	return v.title
}

func (v *resourceView) selChanged(r, c int) {
	v.selectedRow = r
	v.selectItem(r, c)
	v.getTV().cmdBuff.setActive(false)
}

func (v *resourceView) getSelectedItem() string {
	if v.selectedFn != nil {
		return v.selectedFn()
	}
	return v.selectedItem
}

// Protocol...

// Hints fetch menu hints
func (v *resourceView) hints() hints {
	return v.CurrentPage().Item.(hinter).hints()
}

func (v *resourceView) setColorerFn(f colorerFn) {
	v.colorerFn = f
	v.getTV().setColorer(f)
}

func (v *resourceView) setEnterFn(f enterFn) {
	v.enterFn = f
}

func (v *resourceView) setDecorateFn(f decorateFn) {
	v.decorateFn = f
}

// ----------------------------------------------------------------------------
// Actions...

func (v *resourceView) enterCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.flash(flashInfo, "Enter pressed...")
	if v.enterFn != nil {
		v.enterFn(v.app, v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	}
	return nil
}

func (v *resourceView) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.flash(flashInfo, "Refreshing...")
	v.refresh()
	return nil
}

func (v *resourceView) backCmd(*tcell.EventKey) *tcell.EventKey {
	v.switchPage(v.list.GetName())
	return nil
}

func (v *resourceView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	sel := v.getSelectedItem()
	confirm := v.GetPrimitive("confirm").(*tview.Modal)
	confirm.SetText(fmt.Sprintf("Delete %s %s?", v.list.GetName(), sel))
	confirm.SetDoneFunc(func(_ int, button string) {
		if button == "OK" {
			v.getTV().setDeleted()
			v.app.flash(flashInfo, fmt.Sprintf("Deleting %s %s", v.list.GetName(), sel))
			if err := v.list.Resource().Delete(sel); err != nil {
				v.app.flash(flashErr, "Boom!", err.Error())
			} else {
				v.refresh()
			}
		}
		v.switchPage(v.list.GetName())
	})
	v.SwitchToPage("confirm")

	return nil
}

func (v *resourceView) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	sel := v.getSelectedItem()
	raw, err := v.list.Resource().Describe(v.title, sel, v.app.flags)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Warn().Msgf("Describe %v", err.Error())
		return evt
	}
	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("Describe")
		details.setTitle(sel)
		details.SetTextColor(tcell.ColorAqua)
		details.SetText(string(raw))
		details.ScrollToBeginning()
	}
	v.switchPage("details")
	return nil
}

func (v *resourceView) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	sel := v.getSelectedItem()
	raw, err := v.list.Resource().Marshal(sel)
	if err != nil {
		v.app.flash(flashErr, "Unable to marshal resource", err.Error())
		log.Error().Err(err)
		return evt
	}
	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("View")
		details.setTitle(sel)
		details.SetTextColor(tcell.ColorMediumAquamarine)
		details.SetText(string(raw))
		details.ScrollToBeginning()
	}
	v.switchPage("details")
	return nil
}

func (v *resourceView) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	ns, po := namespaced(v.selectedItem)
	args := make([]string, 0, 10)
	args = append(args, "edit")
	args = append(args, v.list.GetName())
	args = append(args, "-n", ns)
	args = append(args, "--context", v.app.config.K9s.CurrentContext)
	args = append(args, po)
	runK(v.app, args...)
	return evt
}

func (v *resourceView) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := v.namespaces[i]
	v.doSwitchNamespace(ns)
	return nil
}

func (v *resourceView) doSwitchNamespace(ns string) {
	v.update.Lock()
	{
		if ns == noSelection {
			ns = resource.AllNamespace
		}
		v.selectedNS = ns
		v.app.flash(flashInfo, fmt.Sprintf("Viewing `%s namespace...", ns))
		v.list.SetNamespace(v.selectedNS)
	}
	v.update.Unlock()
	v.refresh()
	v.selectItem(0, 0)
	v.getTV().resetTitle()
	v.getTV().Select(0, 0)
	v.app.cmdBuff.reset()
	v.app.config.SetActiveNamespace(v.selectedNS)
	v.app.config.Save()
}

func (v *resourceView) refresh() {
	if _, ok := v.CurrentPage().Item.(*tableView); !ok {
		return
	}

	v.update.Lock()
	{
		if v.list.Namespaced() {
			v.list.SetNamespace(v.selectedNS)
		}
		if err := v.list.Reconcile(); err != nil {
			log.Error().Err(err).Msg("Reconciliation failed")
			v.app.flash(flashErr, err.Error())
		}
		data := v.list.Data()
		if v.decorateFn != nil {
			data = v.decorateFn(data)
		}
		v.getTV().update(data)
		v.selectItem(v.selectedRow, 0)
		v.refreshActions()
		v.app.clusterInfoView.refresh()
		v.app.Draw()
	}
	v.update.Unlock()
}

func (v *resourceView) getTV() *tableView {
	if tv, ok := v.GetPrimitive(v.list.GetName()).(*tableView); ok {
		return tv
	}
	return nil
}

func (v *resourceView) selectItem(r, c int) {
	t := v.getTV()
	if r == 0 || t.GetCell(r, 0) == nil {
		v.selectedItem = noSelection
		return
	}

	switch v.list.GetNamespace() {
	case resource.NotNamespaced:
		v.selectedItem = strings.TrimSpace(t.GetCell(r, 0).Text)
	case resource.AllNamespaces:
		v.selectedItem = path.Join(
			strings.TrimSpace(t.GetCell(r, 0).Text),
			strings.TrimSpace(t.GetCell(r, 1).Text),
		)
	default:
		v.selectedItem = path.Join(
			v.selectedNS,
			strings.TrimSpace(t.GetCell(r, 0).Text),
		)
	}
}

func (v *resourceView) switchPage(p string) {
	v.update.Lock()
	{
		v.SwitchToPage(p)
		v.selectedNS = v.list.GetNamespace()
		h := v.GetPrimitive(p).(hinter)
		v.app.setHints(h.hints())
		v.app.SetFocus(v.CurrentPage().Item)
	}
	v.update.Unlock()
}

func (v *resourceView) rowSelected() bool {
	return v.selectedItem != noSelection
}

func namespaced(n string) (string, string) {
	ns, po := path.Split(n)
	return strings.Trim(ns, "/"), po
}

func (v *resourceView) refreshActions() {
	if _, ok := v.CurrentPage().Item.(*tableView); !ok {
		return
	}

	var nn []interface{}
	aa := make(keyActions)
	if k8s.CanIAccess(v.app.conn().Config(), log.Logger, "", "list", "namespaces", "namespace.v1") {
		var err error
		nn, err = k8s.NewNamespace(v.app.conn()).List(resource.AllNamespaces)
		if err != nil {
			log.Warn().Msgf("Access %v", err)
			v.app.flash(flashErr, err.Error())
		}

		if v.list.Namespaced() && !v.list.AllNamespaces() {
			if !config.InNSList(nn, v.list.GetNamespace()) {
				v.list.SetNamespace(resource.DefaultNamespace)
			}
		}

		if v.list.Access(resource.NamespaceAccess) {
			v.namespaces = make(map[int]string, config.MaxFavoritesNS)
			for i, n := range v.app.config.FavNamespaces() {
				aa[tcell.Key(numKeys[i])] = newKeyAction(n, v.switchNamespaceCmd, true)
				v.namespaces[i] = n
			}
		}
	}

	aa[tcell.KeyEnter] = newKeyAction("Enter", v.enterCmd, true)

	aa[tcell.KeyCtrlR] = newKeyAction("Refresh", v.refreshCmd, false)
	aa[KeyHelp] = newKeyAction("Help", v.app.noopCmd, false)
	aa[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	if v.list.Access(resource.EditAccess) {
		aa[KeyE] = newKeyAction("Edit", v.editCmd, true)
	}
	if v.list.Access(resource.DeleteAccess) {
		aa[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, true)
	}
	if v.list.Access(resource.ViewAccess) {
		aa[KeyY] = newKeyAction("YAML", v.viewCmd, true)
	}
	if v.list.Access(resource.DescribeAccess) {
		aa[KeyD] = newKeyAction("Describe", v.describeCmd, true)
	}

	if v.extraActionsFn != nil {
		v.extraActionsFn(aa)
	}
	t := v.getTV()
	t.setActions(aa)
	v.app.setHints(t.hints())
}

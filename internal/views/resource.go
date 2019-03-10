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

const noSelection = ""

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
		namespaces     map[int]string
		selectedNS     string
		update         sync.Mutex
		list           resource.List
		extraActionsFn func(keyActions)
		selectedFn     func() string
		decorateDataFn func(resource.TableData) resource.TableData
	}
)

func newResourceView(title string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := resourceView{
		app:        app,
		title:      title,
		list:       list,
		selectedNS: list.GetNamespace(),
		Pages:      tview.NewPages(),
	}

	tv := newTableView(app, v.title, list.SortFn())
	{
		tv.SetColorer(c)
		tv.SetSelectionChangedFunc(v.selChanged)
	}
	v.AddPage(v.list.GetName(), tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)
	return &v
}

// Init watches all running pods in given namespace
func (v *resourceView) init(ctx context.Context, ns string) {
	v.selectedItem, v.selectedNS = noSelection, ns

	go func(ctx context.Context) {
		initTick := 0.1
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s watcher canceled!", v.title)
				return
			case <-time.After(time.Duration(initTick) * time.Second):
				v.refresh()
				initTick = float64(config.Root.K9s.RefreshRate)
			}
		}
	}(ctx)
	v.refreshActions()
	if tv, ok := v.CurrentPage().Item.(*tableView); ok {
		tv.Select(0, 0)
	}
}

func (v *resourceView) getTitle() string {
	return v.title
}

func (v *resourceView) selChanged(r, c int) {
	v.selectItem(r, c)
	v.getTV().cmdBuff.setActive(false)
}

func (v *resourceView) colorFn(f colorerFn) {
	v.getTV().SetColorer(f)
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

// ----------------------------------------------------------------------------
// Actions...

func (v *resourceView) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	log.Debug().Msg("Refreshing resource...")
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
	v.getTV().setDeleted()
	sel := v.getSelectedItem()
	v.app.flash(flashInfo, fmt.Sprintf("Deleting %s %s", v.list.GetName(), sel))
	if err := v.list.Resource().Delete(sel); err != nil {
		v.app.flash(flashErr, "Boom!", err.Error())
	}
	v.selectedItem = noSelection
	return nil
}

func (v *resourceView) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	sel := v.getSelectedItem()
	raw, err := v.list.Resource().Describe(v.title, sel)
	if err != nil {
		v.app.flash(flashErr, "Unable to describeCmd this resource", err.Error())
		log.Error().Err(err)
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
	ns, s := namespaced(v.selectedItem)
	runK(v.app, "edit", v.list.GetName(), "-n", ns, s)
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
	config.Root.SetActiveNamespace(v.selectedNS)
	config.Root.Save()
}

// Utils...

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
			v.app.flash(flashErr, err.Error())
		}
		data := v.list.Data()
		if v.decorateDataFn != nil {
			data = v.decorateDataFn(data)
		}
		v.getTV().update(data)

		v.refreshActions()
		v.app.infoView.refresh()
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
	if r == 0 {
		v.selectedItem = noSelection
		return
	}

	t := v.getTV()
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

	nn, err := k8s.NewNamespace().List(resource.AllNamespaces)
	if err != nil {
		v.app.flash(flashErr, "Unable to retrieve namespaces", err.Error())
		return
	}

	if v.list.Namespaced() && !v.list.AllNamespaces() {
		if !config.InNSList(nn, v.list.GetNamespace()) {
			v.list.SetNamespace(resource.DefaultNamespace)
		}
	}

	aa := make(keyActions)
	if v.list.Access(resource.NamespaceAccess) {
		v.namespaces = make(map[int]string, config.MaxFavoritesNS)
		for i, n := range config.Root.FavNamespaces() {
			aa[tcell.Key(numKeys[i])] = newKeyAction(n, v.switchNamespaceCmd)
			v.namespaces[i] = n
		}
	}

	aa[tcell.KeyCtrlR] = newKeyAction("Refresh", v.refreshCmd)

	if v.list.Access(resource.EditAccess) {
		aa[KeyE] = newKeyAction("Edit", v.editCmd)
	}

	if v.list.Access(resource.DeleteAccess) {
		aa[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd)
	}
	if v.list.Access(resource.ViewAccess) {
		aa[KeyV] = newKeyAction("View", v.viewCmd)
	}
	if v.list.Access(resource.DescribeAccess) {
		aa[KeyD] = newKeyAction("Describe", v.describeCmd)
	}

	aa[KeyHelp] = newKeyAction("Help", v.app.noopCmd)

	if v.extraActionsFn != nil {
		v.extraActionsFn(aa)
	}

	t := v.getTV()
	t.setActions(aa)
	v.app.setHints(t.hints())
}

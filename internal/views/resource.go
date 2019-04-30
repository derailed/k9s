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
		selectedRow    int
		namespaces     map[int]string
		selectedNS     string
		list           resource.List
		enterFn        enterFn
		extraActionsFn func(keyActions)
		selectedFn     func() string
		decorateFn     decorateFn
		colorerFn      colorerFn
		actions        keyActions
		mx             sync.Mutex
		suspended      bool
	}
)

func newResourceView(title string, app *appView, list resource.List) resourceViewer {
	v := resourceView{
		app:        app,
		title:      title,
		actions:    make(keyActions),
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

	go v.updater(ctx)
	v.refresh()
	if tv, ok := v.CurrentPage().Item.(*tableView); ok {
		tv.Select(1, 0)
		v.selChanged(1, 0)
	}
}

func (v *resourceView) updater(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s watcher canceled!", v.title)
				return
			case <-time.After(time.Duration(v.app.config.K9s.RefreshRate) * time.Second):
				var suspended bool
				v.mx.Lock()
				{
					suspended = v.suspended
				}
				v.mx.Unlock()
				if suspended == true {
					continue
				}
				v.app.QueueUpdate(func() {
					v.refresh()
				})
			}
		}
	}(ctx)
}

func (v *resourceView) suspend() {
	v.mx.Lock()
	{
		v.suspended = true
	}
	v.mx.Unlock()
}

func (v *resourceView) resume() {
	v.mx.Lock()
	{
		v.suspended = false
	}
	v.mx.Unlock()
}

func (v *resourceView) setExtraActionsFn(f actionsFn) {
	f(v.actions)
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

func (v *resourceView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	// If in command mode run filter otherwise enter function.
	if v.getTV().filterCmd(evt) == nil {
		return nil
	}

	if v.enterFn != nil {
		v.enterFn(v.app, v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	} else {
		v.defaultEnter(v.app, v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
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
	v.showModal(fmt.Sprintf("Delete %s %s?", v.list.GetName(), sel), func(_ int, button string) {
		if button == "OK" {
			v.getTV().setDeleted()
			v.app.flash(flashInfo, fmt.Sprintf("Deleting %s %s", v.list.GetName(), sel))
			if err := v.list.Resource().Delete(sel); err != nil {
				v.app.flash(flashErr, "Boom!", err.Error())
			} else {
				v.refresh()
			}
		}
		v.dismissModal()
	})

	return nil
}

func (v *resourceView) showModal(msg string, done func(int, string)) {
	confirm := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(done)
	v.AddPage("confirm", confirm, false, false)
	v.ShowPage("confirm")
}

func (v *resourceView) dismissModal() {
	v.RemovePage("confirm")
	v.switchPage(v.list.GetName())
}

func (v *resourceView) defaultEnter(app *appView, ns, resource, selection string) {
	sel := v.getSelectedItem()
	yaml, err := v.list.Resource().Describe(v.title, sel, v.app.flags)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Warn().Msgf("Describe %v", err.Error())
		return
	}

	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("Describe")
		details.setTitle(sel)
		details.SetTextColor(tcell.ColorAqua)
		details.SetText(colorizeYAML(v.app.styles.Style, yaml))
		details.ScrollToBeginning()
	}
	v.switchPage("details")
}

func (v *resourceView) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.defaultEnter(v.app, v.list.GetNamespace(), v.list.GetName(), v.selectedItem)

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
		details.SetText(colorizeYAML(v.app.styles.Style, raw))
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
	runK(true, v.app, args...)

	return evt
}

func (v *resourceView) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := v.namespaces[i]
	v.doSwitchNamespace(ns)

	return nil
}

func (v *resourceView) doSwitchNamespace(ns string) {
	if ns == "" {
		ns = resource.AllNamespace
	}
	v.selectedNS = ns
	v.app.flash(flashInfo, fmt.Sprintf("Viewing `%s namespace...", ns))
	v.list.SetNamespace(v.selectedNS)

	v.refresh()
	v.getTV().resetTitle()
	v.getTV().Select(1, 0)
	v.selectItem(1, 0)
	v.app.cmdBuff.reset()
	v.app.config.SetActiveNamespace(v.selectedNS)
	v.app.config.Save()
}

func (v *resourceView) refresh() {
	if _, ok := v.CurrentPage().Item.(*tableView); !ok {
		return
	}

	if v.list.Namespaced() {
		v.list.SetNamespace(v.selectedNS)
	}

	v.refreshActions()
	v.app.clusterInfoView.refresh()

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
	v.app.Draw()
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
	v.SwitchToPage(p)
	v.selectedNS = v.list.GetNamespace()
	if h, ok := v.GetPrimitive(p).(hinter); ok {
		v.app.setHints(h.hints())
	}
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
				v.actions[tcell.Key(numKeys[i])] = newKeyAction(n, v.switchNamespaceCmd, true)
				v.namespaces[i] = n
			}
		}
	}

	v.actions[tcell.KeyEnter] = newKeyAction("Enter", v.enterCmd, false)
	v.actions[tcell.KeyCtrlR] = newKeyAction("Refresh", v.refreshCmd, false)
	v.actions[KeyHelp] = newKeyAction("Help", v.app.noopCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	if v.list.Access(resource.EditAccess) {
		v.actions[KeyE] = newKeyAction("Edit", v.editCmd, true)
	}
	if v.list.Access(resource.DeleteAccess) {
		v.actions[tcell.KeyCtrlD] = newKeyAction("Delete", v.deleteCmd, true)
	}
	if v.list.Access(resource.ViewAccess) {
		v.actions[KeyY] = newKeyAction("YAML", v.viewCmd, true)
	}
	if v.list.Access(resource.DescribeAccess) {
		v.actions[KeyD] = newKeyAction("Describe", v.describeCmd, true)
	}

	if v.extraActionsFn != nil {
		v.extraActionsFn(v.actions)
	}
	t := v.getTV()
	t.setActions(v.actions)
	v.app.setHints(t.hints())
}

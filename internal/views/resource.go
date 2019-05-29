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
	noSelection    = ""
	clusterRefresh = time.Duration(15 * time.Second)
)

type updatable interface {
	restartUpdates()
	stopUpdates()
	update(context.Context)
}

type resourceView struct {
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
	// suspended      bool
	nsListAccess bool
	path         *string
	cancelFn     context.CancelFunc
	parentCtx    context.Context
}

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
	tv.SetSelectionChangedFunc(v.selChanged)
	v.AddPage(v.list.GetName(), tv, true, true)

	details := newDetailsView(app, v.backCmd)
	v.AddPage("details", details, true, false)

	return &v
}

func (v *resourceView) stopUpdates() {
	if v.cancelFn != nil {
		log.Debug().Msgf(">>> STOP updates %s", v.list.GetName())
		v.cancelFn()
	}
}

func (v *resourceView) restartUpdates() {
	log.Debug().Msgf(">>> RESTART updates %s", v.list.GetName())
	if v.cancelFn != nil {
		v.cancelFn()
	}

	var vctx context.Context
	vctx, v.cancelFn = context.WithCancel(v.parentCtx)
	v.update(vctx)
}

// Init watches all running pods in given namespace
func (v *resourceView) init(ctx context.Context, ns string) {
	v.parentCtx = ctx
	var vctx context.Context
	vctx, v.cancelFn = context.WithCancel(ctx)
	v.selectedItem, v.selectedNS = noSelection, ns

	colorer := defaultColorer
	if v.colorerFn != nil {
		colorer = v.colorerFn
	}
	v.getTV().setColorer(colorer)

	v.nsListAccess = v.app.conn().CanIAccess("", "namespaces", "namespace.v1", []string{"list"})
	if v.nsListAccess {
		nn, err := k8s.NewNamespace(v.app.conn()).List(resource.AllNamespaces)
		if err != nil {
			log.Warn().Err(err).Msg("List namespaces")
			v.app.flash().errf("Unable to list namespaces %s", err)
		}

		if v.list.Namespaced() && !v.list.AllNamespaces() {
			if !config.InNSList(nn, v.list.GetNamespace()) {
				v.list.SetNamespace(resource.DefaultNamespace)
			}
		}
	}

	v.update(vctx)
	v.app.clusterInfoView.refresh()
	v.refresh()
	if tv, ok := v.CurrentPage().Item.(*tableView); ok {
		tv.Select(1, 0)
	}
}

func (v *resourceView) update(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s cluster updater canceled!", v.list.GetName())
				return
			case <-time.After(clusterRefresh):
				v.app.QueueUpdateDraw(func() {
					v.app.clusterInfoView.refresh()
				})
			}
		}
	}(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s updater canceled!", v.list.GetName())
				return
			case <-time.After(time.Duration(v.app.config.K9s.RefreshRate) * time.Second):
				v.app.QueueUpdateDraw(func() {
					log.Debug().Msgf(">>> Refreshing %s", v.list.GetName())
					v.refresh()
				})
			}
		}
	}(ctx)
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
		v.defaultEnter(v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	}

	return nil
}

func (v *resourceView) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.flash().info("Refreshing...")
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
			v.app.flash().infof("Deleting %s %s", v.list.GetName(), sel)
			if err := v.list.Resource().Delete(sel); err != nil {
				v.app.flash().errf("Delete failed with %s", err)
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

func (v *resourceView) defaultEnter(ns, resource, selection string) {
	yaml, err := v.list.Resource().Describe(v.title, selection, v.app.flags)
	if err != nil {
		v.app.flash().errf("Describe command failed %s", err)
		log.Warn().Msgf("Describe %v", err.Error())
		return
	}

	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("Describe")
		details.setTitle(selection)
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

	log.Debug().Msgf("Selected Item %v-%v-%v", v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	v.defaultEnter(v.list.GetNamespace(), v.list.GetName(), v.selectedItem)

	return nil
}

func (v *resourceView) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	sel := v.getSelectedItem()
	raw, err := v.list.Resource().Marshal(sel)
	if err != nil {
		v.app.flash().errf("Unable to marshal resource %s", err)
		log.Error().Err(err)
		return evt
	}
	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("YAML")
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

	v.stopUpdates()
	{
		ns, po := namespaced(v.selectedItem)
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, v.list.GetName())
		args = append(args, "-n", ns)
		args = append(args, "--context", v.app.config.K9s.CurrentContext)
		args = append(args, po)
		runK(true, v.app, args...)
	}
	v.restartUpdates()

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
	v.app.flash().infof("Viewing `%s namespace...", ns)
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

	v.refreshActions()

	if v.list.Namespaced() {
		v.list.SetNamespace(v.selectedNS)
	}
	if err := v.list.Reconcile(v.app.informer, v.path); err != nil {
		log.Error().Err(err).Msgf("Reconciliation for %s failed", v.title)
		v.app.flash().errf("Reconciliation for %s failed - %s", v.title, err)
	}
	data := v.list.Data()
	if v.decorateFn != nil {
		data = v.decorateFn(data)
	}
	v.getTV().update(data)
	v.selectItem(v.selectedRow, 0)
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

	col0 := strings.TrimSpace(t.GetCell(r, 0).Text)
	switch v.list.GetNamespace() {
	case resource.NotNamespaced:
		v.selectedItem = col0
	case resource.AllNamespaces:
		v.selectedItem = path.Join(col0, strings.TrimSpace(t.GetCell(r, 1).Text))
	default:
		v.selectedItem = path.Join(v.selectedNS, col0)
	}
}

func (v *resourceView) switchPage(p string) {
	log.Debug().Msgf("Switching page to %s", p)
	if _, ok := v.CurrentPage().Item.(*tableView); ok {
		v.stopUpdates()
	} else {
		log.Debug().Msgf("Not a table %T", v.CurrentPage().Item)
	}

	v.SwitchToPage(p)
	v.selectedNS = v.list.GetNamespace()
	if vu, ok := v.GetPrimitive(p).(hinter); ok {
		v.app.setHints(vu.hints())
	}

	if _, ok := v.CurrentPage().Item.(*tableView); ok {
		v.restartUpdates()
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
	if v.list.Access(resource.NamespaceAccess) {
		v.namespaces = make(map[int]string, config.MaxFavoritesNS)
		v.actions[tcell.Key(numKeys[0])] = newKeyAction(resource.AllNamespace, v.switchNamespaceCmd, true)
		v.namespaces[0] = resource.AllNamespace
		index := 1
		for _, n := range v.app.config.FavNamespaces() {
			if n == resource.AllNamespace {
				continue
			}
			v.actions[tcell.Key(numKeys[index])] = newKeyAction(n, v.switchNamespaceCmd, true)
			v.namespaces[index] = n
			index++
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

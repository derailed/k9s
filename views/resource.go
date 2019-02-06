package views

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
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
		tv.table.SetSelectionChangedFunc(v.selChanged)
	}
	v.AddPage(v.list.GetName(), tv, true, true)

	var xray details
	if list.HasXRay() {
		xray = newXrayView(app)
	} else {
		xray = newYamlView(app)
	}
	xray.setActions(keyActions{
		tcell.KeyEscape: {description: "Back", action: v.back},
	})

	details := newDetailsView()
	details.setActions(keyActions{
		tcell.KeyEscape: {description: "Back", action: v.back},
	})

	v.AddPage("details", details, true, false)
	v.AddPage("xray", xray, true, false)

	return &v
}

// Init watches all running pods in given namespace
func (v *resourceView) init(ctx context.Context, ns string) {
	details := v.GetPrimitive("xray").(details)
	details.clear()

	v.selectedItem, v.selectedNS = noSelection, ns

	go func(ctx context.Context) {
		initTick := 0.1
		for {
			select {
			case <-ctx.Done():
				log.Debugf("%s watcher canceled!", v.title)
				return
			case <-time.After(time.Duration(initTick) * time.Second):
				v.refresh()
				initTick = float64(k9sCfg.K9s.RefreshRate)
			}
		}
	}(ctx)
	v.refreshActions()
	v.CurrentPage().Item.(*tableView).table.Select(0, 0)
}

func (v *resourceView) selChanged(r, c int) {
	v.selectItem(r, c)
}

func (v *resourceView) colorFn(f colorerFn) {
	v.getTV().SetColorer(f)
}

// Protocol...

// Hints fetch menu hints
func (v *resourceView) hints() hints {
	return v.CurrentPage().Item.(hinter).hints()
}

// Actions...

func (v *resourceView) back(*tcell.EventKey) {
	v.switchPage(v.list.GetName())
}

func (v *resourceView) delete(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	v.getTV().setDeleted()
	v.app.flash(flashInfo, fmt.Sprintf("Deleting %s %s", v.list.GetName(), v.selectedItem))
	if err := v.list.Resource().Delete(v.selectedItem); err != nil {
		v.app.flash(flashErr, "Boom!", err.Error())
	}
	v.selectedItem = noSelection
}

func (v *resourceView) describe(*tcell.EventKey) {
	details := v.GetPrimitive("xray").(details)
	details.clear()

	if !v.rowSelected() {
		return
	}

	props, err := v.list.Describe(v.selectedItem)
	if err != nil {
		v.app.flash(flashErr, "Unable to get xray fields", err.Error())
		return
	}
	details.update(props)
	details.setTitle(fmt.Sprintf(" %s ", v.selectedItem))
	v.switchPage("xray")
}

func (v *resourceView) view(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	raw, err := v.list.Resource().Marshal(v.selectedItem)
	if err != nil {
		v.app.flash(flashErr, "Unable to marshal resource", err.Error())
		log.Error(err)
		return
	}

	var re = regexp.MustCompile(`([\w|\.|"|\-|\/|\@]+):(.*)`)
	str := re.ReplaceAllString(string(raw), `[aqua]$1: [white]$2`)

	details := v.GetPrimitive("details").(*detailsView)
	details.SetText(str)
	details.setTitle(v.selectedItem)
	v.switchPage("details")
}

func (v *resourceView) edit(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}
	v.app.flash(flashInfo, fmt.Sprintf("Editing %s %s", v.title, v.selectedItem))
	ns, s := namespaced(v.selectedItem)
	run(v.app, "edit", v.list.GetName(), "-n", ns, s)
	return
}

func (v *resourceView) switchNamespace(evt *tcell.EventKey) {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := v.namespaces[i]
	v.doSwitchNamespace(ns)
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
	v.getTV().table.Select(0, 0)
	v.app.cmdBuff.reset()
	k9sCfg.K9s.Namespace.Active = v.selectedNS
	k9sCfg.validateAndSave()
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
	return v.GetPrimitive(v.list.GetName()).(*tableView)
}

func (v *resourceView) getSelectedItem() string {
	return v.selectedItem
}

func (v *resourceView) selectItem(r, c int) {
	if r == 0 {
		v.selectedItem = noSelection
		return
	}

	t := v.getTV()
	switch v.list.GetNamespace() {
	case resource.NotNamespaced:
		v.selectedItem = strings.TrimSpace(t.table.GetCell(r, 0).Text)
	case resource.AllNamespaces:
		v.selectedItem = path.Join(
			strings.TrimSpace(t.table.GetCell(r, 0).Text),
			strings.TrimSpace(t.table.GetCell(r, 1).Text),
		)
	default:
		v.selectedItem = path.Join(
			v.selectedNS,
			strings.TrimSpace(t.table.GetCell(r, 0).Text),
		)
	}
}

func (v *resourceView) switchPage(p string) {
	v.update.Lock()
	{
		v.SwitchToPage(p)
		h := v.GetPrimitive(p).(hinter)
		v.selectedNS = v.list.GetNamespace()
		v.app.setHints(h.hints())
		v.app.SetFocus(v.CurrentPage().Item)
	}
	v.update.Unlock()
}

func (v *resourceView) rowSelected() bool {
	item := v.getSelectedItem()
	return item != noSelection
}

func namespaced(n string) (string, string) {
	ns, po := path.Split(n)
	return strings.Trim(ns, "/"), po
}

func (v *resourceView) refreshActions() {
	if _, ok := v.CurrentPage().Item.(*tableView); !ok {
		return
	}

	nn, err := k8s.NewNamespace().List(defaultNS)
	if err != nil {
		v.app.flash(flashErr, "Unable to retrieve namespaces", err.Error())
		return
	}

	if v.list.Namespaced() && !v.list.AllNamespaces() {
		if !inNSList(nn, v.list.GetNamespace()) {
			v.list.SetNamespace(resource.DefaultNamespace)
		}
	}

	aa := keyActions{}
	if v.list.Access(resource.NamespaceAccess) {
		v.namespaces = make(map[int]string, maxFavorites)
		var i int
		for _, n := range k9sCfg.K9s.Namespace.Favorites {
			if i > maxFavorites {
				break
			}

			if n == resource.AllNamespace {
				aa[tcell.Key(numKeys[i])] = newKeyHandler(resource.AllNamespace, v.switchNamespace)
				v.namespaces[i] = resource.AllNamespaces
				i++
				continue
			}

			if inNSList(nn, n) {
				aa[tcell.Key(numKeys[i])] = newKeyHandler(n, v.switchNamespace)
				v.namespaces[i] = n
				i++
			} else {
				k9sCfg.rmFavNS(n)
				k9sCfg.validateAndSave()
			}
		}
	}

	if v.list.Access(resource.EditAccess) {
		aa[tcell.KeyCtrlE] = newKeyHandler("Edit", v.edit)
	}

	if v.list.Access(resource.DeleteAccess) {
		aa[tcell.KeyBackspace2] = newKeyHandler("Delete", v.delete)
	}
	if v.list.Access(resource.ViewAccess) {
		aa[tcell.KeyCtrlV] = newKeyHandler("View", v.view)
	}
	if v.list.Access(resource.DescribeAccess) {
		aa[tcell.KeyCtrlX] = newKeyHandler("Describe", v.describe)
	}

	aa[tcell.KeyCtrlQ] = newKeyHandler("Quit", v.app.quit)
	aa[tcell.KeyCtrlA] = newKeyHandler("Aliases", v.app.help)

	if v.extraActionsFn != nil {
		v.extraActionsFn(aa)
	}

	t := v.getTV()
	t.setActions(aa)
	v.app.setHints(t.hints())
}

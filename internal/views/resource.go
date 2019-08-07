package views

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	updatable interface {
		restartUpdates()
		stopUpdates()
		update(context.Context)
	}

	resourceView struct {
		*masterDetail

		namespaces map[int]string
		list       resource.List
		cancelFn   context.CancelFunc
		parentCtx  context.Context
		path       *string
		colorerFn  ui.ColorerFunc
		decorateFn decorateFn
	}
)

func newResourceView(title string, app *appView, list resource.List) resourceViewer {
	v := resourceView{list: list}
	v.masterDetail = newMasterDetail(title, list.GetNamespace(), app, v.backCmd)

	return &v
}

// Init watches all running pods in given namespace
func (v *resourceView) Init(ctx context.Context, ns string) {
	v.masterDetail.init(ctx, ns)
	v.masterPage().setFilterFn(v.filterResource)
	if v.colorerFn != nil {
		v.masterPage().SetColorerFn(v.colorerFn)
	}

	v.parentCtx = ctx
	var vctx context.Context
	vctx, v.cancelFn = context.WithCancel(ctx)

	colorer := ui.DefaultColorer
	if v.colorerFn != nil {
		colorer = v.colorerFn
	}
	v.masterPage().SetColorerFn(colorer)

	v.update(vctx)
	v.app.clusterInfo().refresh()
	v.refresh()

	tv := v.masterPage()
	r, _ := tv.GetSelection()
	if r == 0 && tv.GetRowCount() > 0 {
		tv.Select(1, 0)
	}
}

func (v *resourceView) setColorerFn(f ui.ColorerFunc) {
	v.colorerFn = f
}

func (v *resourceView) setDecorateFn(f decorateFn) {
	v.decorateFn = f
}

func (v *resourceView) filterResource(sel string) {
	v.list.SetLabelSelector(sel)
	v.refresh()
}

func (v *resourceView) stopUpdates() {
	if v.cancelFn != nil {
		v.cancelFn()
	}
}

func (v *resourceView) restartUpdates() {
	if v.cancelFn != nil {
		v.cancelFn()
	}

	var vctx context.Context
	vctx, v.cancelFn = context.WithCancel(v.parentCtx)
	v.update(vctx)
}

func (v *resourceView) update(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s updater canceled!", v.list.GetName())
				return
			case <-time.After(time.Duration(v.app.Config.K9s.GetRefreshRate()) * time.Second):
				v.app.QueueUpdateDraw(func() {
					v.refresh()
				})
			}
		}
	}(ctx)
}

func (v *resourceView) backCmd(*tcell.EventKey) *tcell.EventKey {
	v.switchPage("master")
	return nil
}

func (v *resourceView) switchPage(p string) {
	log.Debug().Msgf("Switching page to %s", p)
	if _, ok := v.CurrentPage().Item.(*tableView); ok {
		v.stopUpdates()
	}

	v.SwitchToPage(p)
	v.currentNS = v.list.GetNamespace()
	if vu, ok := v.GetPrimitive(p).(ui.Hinter); ok {
		v.app.SetHints(vu.Hints())
	}

	if _, ok := v.CurrentPage().Item.(*tableView); ok {
		v.restartUpdates()
	}
}

// ----------------------------------------------------------------------------
// Actions...

func (v *resourceView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msg("ResourceV Got Enter")
	// If in command mode run filter otherwise enter function.
	if v.masterPage().filterCmd(evt) == nil || !v.rowSelected() {
		return nil
	}

	if v.enterFn != nil {
		log.Debug().Msg("Custom enter!!")
		v.enterFn(v.app, v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	} else {
		v.defaultEnter(v.list.GetNamespace(), v.list.GetName(), v.selectedItem)
	}

	return nil
}

func (v *resourceView) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.Flash().Info("Refreshing...")
	v.refresh()
	return nil
}

func (v *resourceView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	sel := v.getSelectedItem()
	msg := fmt.Sprintf("Delete %s %s?", v.list.GetName(), sel)
	dialog.ShowDelete(v.Pages, msg, func(cascade, force bool) {
		v.masterPage().ShowDeleted()
		v.app.Flash().Infof("Delete resource %s %s", v.list.GetName(), sel)
		if err := v.list.Resource().Delete(sel, cascade, force); err != nil {
			v.app.Flash().Errf("Delete failed with %s", err)
		} else {
			v.refresh()
		}
	}, func() {
		v.switchPage("master")
	})
	return nil
}

func (v *resourceView) defaultEnter(ns, _, selection string) {
	if !v.list.Access(resource.DescribeAccess) {
		return
	}

	yaml, err := v.list.Resource().Describe(v.masterPage().GetBaseTitle(), selection)
	if err != nil {
		v.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := v.detailsPage()
	details.setCategory("Describe")
	details.setTitle(selection)
	details.SetTextColor(v.app.Styles.FgColor())
	details.SetText(colorizeYAML(v.app.Styles.Views().Yaml, yaml))
	details.ScrollToBeginning()

	v.switchPage("details")
}

func (v *resourceView) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

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
		v.app.Flash().Errf("Unable to marshal resource %s", err)
		return evt
	}
	details := v.detailsPage()
	details.setCategory("YAML")
	details.setTitle(sel)
	details.SetTextColor(v.app.Styles.FgColor())
	details.SetText(colorizeYAML(v.app.Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	v.switchPage("details")

	return nil
}

func (v *resourceView) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	kcfg := v.app.Conn().Config().Flags().KubeConfig
	v.stopUpdates()
	{
		ns, po := namespaced(v.selectedItem)
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, v.list.GetName())
		args = append(args, "-n", ns)
		args = append(args, "--context", v.app.Config.K9s.CurrentContext)
		if kcfg != nil && *kcfg != "" {
			args = append(args, "--kubeconfig", *kcfg)
		}
		runK(true, v.app, append(args, po)...)
	}
	v.restartUpdates()

	return evt
}

func (v *resourceView) setNamespace(ns string) {
	if v.list.Namespaced() {
		v.currentNS = ns
		v.list.SetNamespace(ns)
	}
}

func (v *resourceView) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := v.namespaces[i]
	if ns == "" {
		ns = resource.AllNamespace
	}
	if v.currentNS == ns {
		return nil
	}

	v.setNamespace(ns)
	v.app.Flash().Infof("Viewing `%s namespace...", ns)
	v.refresh()
	v.masterPage().UpdateTitle()
	v.masterPage().Select(1, 0)
	v.selectItem(1, 0)
	v.app.GetCmdBuff().Reset()
	v.app.Config.SetActiveNamespace(v.currentNS)
	v.app.Config.Save()

	return nil
}

func (v *resourceView) refresh() {
	if _, ok := v.CurrentPage().Item.(*tableView); !ok {
		return
	}

	v.refreshActions()
	if v.list.Namespaced() {
		v.list.SetNamespace(v.currentNS)
	}
	if err := v.list.Reconcile(v.app.informer, v.path); err != nil {
		v.app.Flash().Errf("Reconciliation for %s failed - %s", v.list.GetName(), err)
	}
	data := v.list.Data()
	if v.decorateFn != nil {
		data = v.decorateFn(data)
	}
	v.masterPage().Update(data)
	v.selectItem(v.selectedRow, 0)
}

func (v *resourceView) namespaceActions() {
	if !v.list.Access(resource.NamespaceAccess) {
		return
	}
	v.namespaces = make(map[int]string, config.MaxFavoritesNS)
	// User can't list namespace. Don't offer a choice.
	if v.app.Conn().CheckListNSAccess() != nil {
		return
	}
	v.actions[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(resource.AllNamespace, v.switchNamespaceCmd, true)
	v.namespaces[0] = resource.AllNamespace
	index := 1
	for _, n := range v.app.Config.FavNamespaces() {
		if n == resource.AllNamespace {
			continue
		}
		v.actions[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(n, v.switchNamespaceCmd, true)
		v.namespaces[index] = n
		index++
	}
}

func (v *resourceView) refreshActions() {
	v.namespaceActions()
	v.defaultActions()

	v.actions[tcell.KeyEnter] = ui.NewKeyAction("Enter", v.enterCmd, false)
	v.actions[tcell.KeyCtrlR] = ui.NewKeyAction("Refresh", v.refreshCmd, false)

	if v.list.Access(resource.EditAccess) {
		v.actions[ui.KeyE] = ui.NewKeyAction("Edit", v.editCmd, true)
	}
	if v.list.Access(resource.DeleteAccess) {
		v.actions[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", v.deleteCmd, true)
	}
	if v.list.Access(resource.ViewAccess) {
		v.actions[ui.KeyY] = ui.NewKeyAction("YAML", v.viewCmd, true)
	}
	if v.list.Access(resource.DescribeAccess) {
		v.actions[ui.KeyD] = ui.NewKeyAction("Describe", v.describeCmd, true)
	}

	t := v.masterPage()
	t.SetActions(v.actions)
	v.app.SetHints(t.Hints())
}

package views

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// EnvFn represent the current view exposed environment.
type envFn func() K9sEnv

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
		envFn      envFn
		gvr        string
	}
)

func newResourceView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := resourceView{
		list: list,
		gvr:  gvr,
	}
	v.masterDetail = newMasterDetail(title, list.GetNamespace(), app, v.backCmd)
	v.envFn = v.defaultK9sEnv

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

func (v *resourceView) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	_, n := namespaced(v.masterPage().GetSelectedItem())
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	v.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		v.app.Flash().Err(err)
	}

	return nil
}

func (v *resourceView) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	// If in command mode run filter otherwise enter function.
	if v.masterPage().filterCmd(evt) == nil || !v.masterPage().RowSelected() {
		return nil
	}

	f := v.defaultEnter
	if v.enterFn != nil {
		f = v.enterFn
	}
	f(v.app, v.list.GetNamespace(), v.list.GetName(), v.masterPage().GetSelectedItem())

	return nil
}

func (v *resourceView) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.Flash().Info("Refreshing...")
	v.refresh()
	return nil
}

func (v *resourceView) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItems()
	var msg string
	if len(sel) > 1 {
		msg = fmt.Sprintf("Delete %d selected %s?", len(sel), v.list.GetName())
	} else {
		msg = fmt.Sprintf("Delete %s %s?", v.list.GetName(), sel[0])
	}
	dialog.ShowDelete(v.Pages, msg, func(cascade, force bool) {
		v.masterPage().ShowDeleted()
		if len(sel) > 1 {
			v.app.Flash().Infof("Delete %d selected %s", len(sel), v.list.GetName())
		} else {
			v.app.Flash().Infof("Delete resource %s %s", v.list.GetName(), sel[0])
		}
		for _, res := range sel {
			if err := v.list.Resource().Delete(res, cascade, force); err != nil {
				v.app.Flash().Errf("Delete failed with %s", err)
			} else {
				deletePortForward(v.app.forwarders, res)
			}
		}
		v.refresh()
	}, func() {
		v.switchPage("master")
	})
	return nil
}

func (v *resourceView) markCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	v.masterPage().ToggleMark()
	v.refresh()
	v.app.Draw()
	return nil
}

func deletePortForward(ff map[string]forwarder, sel string) {
	for k, v := range ff {
		tokens := strings.Split(k, ":")
		if tokens[0] == sel {
			log.Debug().Msgf("Deleting associated portForward %s", k)
			v.Stop()
		}
	}
}

func (v *resourceView) defaultEnter(app *appView, ns, _, selection string) {
	if !v.list.Access(resource.DescribeAccess) {
		return
	}

	yaml, err := v.list.Resource().Describe(v.gvr, selection)
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
	if !v.masterPage().RowSelected() {
		return evt
	}
	v.defaultEnter(v.app, v.list.GetNamespace(), v.list.GetName(), v.masterPage().GetSelectedItem())

	return nil
}

func (v *resourceView) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItem()
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
	if !v.masterPage().RowSelected() {
		return evt
	}

	v.stopUpdates()
	{
		ns, po := namespaced(v.masterPage().GetSelectedItem())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, v.list.GetName())
		args = append(args, "-n", ns)
		args = append(args, "--context", v.app.Config.K9s.CurrentContext)
		if cfg := v.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
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

	v.app.switchNS(ns)
	v.setNamespace(ns)
	v.app.Flash().Infof("Viewing namespace `%s`...", ns)
	v.refresh()
	v.masterPage().UpdateTitle()
	v.masterPage().SelectRow(1, true)
	v.app.CmdBuff().Reset()
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
		v.app.Flash().Err(err)
	}
	data := v.list.Data()
	if v.decorateFn != nil {
		data = v.decorateFn(data)
	}
	v.masterPage().Update(data)
}

func (v *resourceView) namespaceActions(aa ui.KeyActions) {
	if !v.list.Access(resource.NamespaceAccess) {
		return
	}
	v.namespaces = make(map[int]string, config.MaxFavoritesNS)
	// User can't list namespace. Don't offer a choice.
	if v.app.Conn().CheckListNSAccess() != nil {
		return
	}
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(resource.AllNamespace, v.switchNamespaceCmd, true)
	v.namespaces[0] = resource.AllNamespace
	index := 1
	for _, n := range v.app.Config.FavNamespaces() {
		if n == resource.AllNamespace {
			continue
		}
		aa[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(n, v.switchNamespaceCmd, true)
		v.namespaces[index] = n
		index++
	}
}

func (v *resourceView) refreshActions() {
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", v.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Enter", v.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", v.refreshCmd, false),
	}
	aa[ui.KeySpace] = ui.NewKeyAction("Mark", v.markCmd, true)
	v.namespaceActions(aa)
	v.defaultActions(aa)

	if v.list.Access(resource.EditAccess) {
		aa[ui.KeyE] = ui.NewKeyAction("Edit", v.editCmd, true)
	}
	if v.list.Access(resource.DeleteAccess) {
		aa[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", v.deleteCmd, true)
	}
	if v.list.Access(resource.ViewAccess) {
		aa[ui.KeyY] = ui.NewKeyAction("YAML", v.viewCmd, true)
	}
	if v.list.Access(resource.DescribeAccess) {
		aa[ui.KeyD] = ui.NewKeyAction("Describe", v.describeCmd, true)
	}
	v.customActions(aa)

	t := v.masterPage()
	t.SetActions(aa)
	v.app.SetHints(t.Hints())
}

func (v *resourceView) customActions(aa ui.KeyActions) {
	pp := config.NewPlugins()
	if err := pp.Load(); err != nil {
		log.Warn().Msgf("No plugin configuration found")
		return
	}

	for k, plugin := range pp.Plugin {
		if !in(plugin.Scopes, v.list.GetName()) {
			continue
		}
		key, err := asKey(plugin.ShortCut)
		if err != nil {
			log.Error().Err(err).Msg("Unable to map shortcut to a key")
			continue
		}
		_, ok := aa[key]
		if ok {
			log.Error().Err(fmt.Errorf("Doh! you are trying to overide an existing command `%s", k)).Msg("Invalid shortcut")
			continue
		}
		aa[key] = ui.NewKeyAction(
			plugin.Description,
			v.execCmd(plugin.Command, plugin.Background, plugin.Args...),
			true)
	}
}

func (v *resourceView) execCmd(bin string, bg bool, args ...string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if !v.masterPage().RowSelected() {
			return evt
		}

		var (
			env = v.envFn()
			aa  = make([]string, len(args))
			err error
		)
		for i, a := range args {
			aa[i], err = env.envFor(a)
			if err != nil {
				log.Error().Err(err).Msg("Args match failed")
				return nil
			}
		}

		if run(true, v.app, bin, bg, aa...) {
			v.app.Flash().Info("Custom CMD launched!")
		} else {
			v.app.Flash().Info("Custom CMD failed!")
		}
		return nil
	}
}

func (v *resourceView) defaultK9sEnv() K9sEnv {
	ns, n := namespaced(v.masterPage().GetSelectedItem())
	ctx, err := v.app.Conn().Config().CurrentContextName()
	if err != nil {
		ctx = "n/a"
	}
	cluster, err := v.app.Conn().Config().CurrentClusterName()
	if err != nil {
		cluster = "n/a"
	}
	user, err := v.app.Conn().Config().CurrentUserName()
	if err != nil {
		user = "n/a"
	}
	groups, err := v.app.Conn().Config().CurrentGroupNames()
	if err != nil {
		groups = []string{"n/a"}
	}
	var cfg string
	kcfg := v.app.Conn().Config().Flags().KubeConfig
	if kcfg != nil && *kcfg != "" {
		cfg = *kcfg
	}

	env := K9sEnv{
		"NAMESPACE":  ns,
		"NAME":       n,
		"CONTEXT":    ctx,
		"CLUSTER":    cluster,
		"USER":       user,
		"GROUPS":     strings.Join(groups, ","),
		"KUBECONFIG": cfg,
	}

	row := v.masterPage().GetRow()
	for i, r := range row {
		env["COL"+strconv.Itoa(i)] = r
	}

	return env
}

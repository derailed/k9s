package view

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// EnvFn represent the current view exposed environment.
type envFn func() K9sEnv

// Resource represents a generic resource viewer.
type Resource struct {
	*MasterDetail

	namespaces map[int]string
	list       resource.List
	cancelFn   context.CancelFunc
	path       *string
	colorerFn  ui.ColorerFunc
	decorateFn decorateFn
	envFn      envFn
	gvr        string
	currentNS  string
}

// NewResource returns a new viewer.
func NewResource(title, gvr string, list resource.List) *Resource {
	return &Resource{
		MasterDetail: NewMasterDetail(),
		list:         list,
		gvr:          gvr,
	}
}

// Init watches all running pods in given namespace
func (r *Resource) Init(ctx context.Context) {
	r.MasterDetail.Init(ctx)
	r.envFn = r.defaultK9sEnv

	table := r.masterPage()
	table.setFilterFn(r.filterResource)
	colorer := ui.DefaultColorer
	if r.colorerFn != nil {
		colorer = r.colorerFn
	}
	table.SetColorerFn(colorer)
	row, _ := table.GetSelection()
	if row == 0 && table.GetRowCount() > 0 {
		table.Select(1, 0)
	}
	r.DumpPages()

	r.refresh()
}

// Start initializes updates.
func (r *Resource) Start() {
	r.Stop()
	var ctx context.Context
	ctx, r.cancelFn = context.WithCancel(context.Background())
	r.update(ctx)
}

// Stop terminates updates.
func (r *Resource) Stop() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}

// Name returns the component name.
func (r *Resource) Name() string {
	return r.list.GetName()
}

// Hints returns the current viewer hints
func (r *Resource) Hints() model.MenuHints {
	if r.CurrentPage() == nil {
		return nil
	}
	if c, ok := r.CurrentPage().Item.(model.Hinter); ok {
		return c.Hints()
	}

	return nil
}

func (r *Resource) setColorerFn(f ui.ColorerFunc) {
	r.colorerFn = f
}

func (r *Resource) setDecorateFn(f decorateFn) {
	r.decorateFn = f
}

func (r *Resource) filterResource(sel string) {
	r.list.SetLabelSelector(sel)
	r.refresh()
}

func (r *Resource) update(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("%s updater canceled!", r.list.GetName())
				return
			case <-time.After(time.Duration(r.app.Config.K9s.GetRefreshRate()) * time.Second):
				r.app.QueueUpdateDraw(func() {
					r.refresh()
				})
			}
		}
	}(ctx)
}

func (r *Resource) backCmd(*tcell.EventKey) *tcell.EventKey {
	r.switchPage("master")
	return nil
}

func (r *Resource) switchPage(p string) {
	log.Debug().Msgf("Switching page to %s", p)
	if _, ok := r.CurrentPage().Item.(*Table); ok {
		r.Stop()
	}

	r.SwitchToPage(p)

	if _, ok := r.CurrentPage().Item.(*Table); ok {
		r.Start()
	}
}

// ----------------------------------------------------------------------------
// Actions...

func (r *Resource) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	_, n := namespaced(r.masterPage().GetSelectedItem())
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	r.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		r.app.Flash().Err(err)
	}

	return nil
}

func (r *Resource) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	// If in command mode run filter otherwise enter function.
	if r.masterPage().filterCmd(evt) == nil || !r.masterPage().RowSelected() {
		return nil
	}

	f := r.defaultEnter
	if r.enterFn != nil {
		f = r.enterFn
	}
	f(r.app, r.list.GetNamespace(), r.list.GetName(), r.masterPage().GetSelectedItem())

	return nil
}

func (r *Resource) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	r.app.Flash().Info("Refreshing...")
	r.refresh()
	return nil
}

func (r *Resource) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	sel := r.masterPage().GetSelectedItems()
	var msg string
	if len(sel) > 1 {
		msg = fmt.Sprintf("Delete %d selected %s?", len(sel), r.list.GetName())
	} else {
		msg = fmt.Sprintf("Delete %s %s?", r.list.GetName(), sel[0])
	}
	dialog.ShowDelete(r.Pages, msg, func(cascade, force bool) {
		r.masterPage().ShowDeleted()
		if len(sel) > 1 {
			r.app.Flash().Infof("Delete %d selected %s", len(sel), r.list.GetName())
		} else {
			r.app.Flash().Infof("Delete resource %s %s", r.list.GetName(), sel[0])
		}
		for _, res := range sel {
			if err := r.list.Resource().Delete(res, cascade, force); err != nil {
				r.app.Flash().Errf("Delete failed with %s", err)
			} else {
				deletePortForward(r.app.forwarders, res)
			}
		}
		r.refresh()
	}, func() {
		r.switchPage("master")
	})
	return nil
}

func (r *Resource) markCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}
	r.masterPage().ToggleMark()
	r.refresh()
	r.app.Draw()

	return nil
}

func deletePortForward(ff map[string]forwarder, sel string) {
	for k, f := range ff {
		tokens := strings.Split(k, ":")
		if tokens[0] == sel {
			log.Debug().Msgf("Deleting associated portForward %s", k)
			f.Stop()
		}
	}
}

func (r *Resource) defaultEnter(app *App, ns, _, selection string) {
	if !r.list.Access(resource.DescribeAccess) {
		return
	}

	yaml, err := r.list.Resource().Describe(r.gvr, selection)
	if err != nil {
		r.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := r.detailsPage()
	details.setCategory("Describe")
	details.setTitle(selection)
	details.SetTextColor(r.app.Styles.FgColor())
	details.SetText(colorizeYAML(r.app.Styles.Views().Yaml, yaml))
	details.ScrollToBeginning()
	r.switchPage("details")
}

func (r *Resource) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}
	r.defaultEnter(r.app, r.list.GetNamespace(), r.list.GetName(), r.masterPage().GetSelectedItem())

	return nil
}

func (r *Resource) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	sel := r.masterPage().GetSelectedItem()
	raw, err := r.list.Resource().Marshal(sel)
	if err != nil {
		r.app.Flash().Errf("Unable to marshal resource %s", err)
		return evt
	}
	details := r.detailsPage()
	details.setCategory("YAML")
	details.setTitle(sel)
	details.SetTextColor(r.app.Styles.FgColor())
	details.SetText(colorizeYAML(r.app.Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	r.app.Content.Push(details)

	return nil
}

func (r *Resource) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	r.Stop()
	{
		ns, po := namespaced(r.masterPage().GetSelectedItem())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, r.list.GetName())
		args = append(args, "-n", ns)
		args = append(args, "--context", r.app.Config.K9s.CurrentContext)
		if cfg := r.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		runK(true, r.app, append(args, po)...)
	}
	r.Start()

	return evt
}

func (r *Resource) setNamespace(ns string) {
	if r.list.Namespaced() {
		r.list.SetNamespace(ns)
	}
}

func (r *Resource) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := r.namespaces[i]
	if ns == "" {
		ns = resource.AllNamespace
	}
	if r.currentNS == ns {
		return nil
	}

	r.app.switchNS(ns)
	r.setNamespace(ns)
	r.app.Flash().Infof("Viewing namespace `%s`...", ns)
	r.refresh()
	r.masterPage().UpdateTitle()
	r.masterPage().SelectRow(1, true)
	r.app.CmdBuff().Reset()
	if err := r.app.Config.SetActiveNamespace(r.currentNS); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	r.app.Config.Save()

	return nil
}

func (r *Resource) refresh() {
	if r.CurrentPage() == nil {
		return
	}
	if _, ok := r.CurrentPage().Item.(*Table); !ok {
		return
	}

	r.refreshActions()
	if r.list.Namespaced() {
		r.list.SetNamespace(r.currentNS)
	}
	if err := r.list.Reconcile(r.app.informer, r.path); err != nil {
		r.app.Flash().Err(err)
	}
	data := r.list.Data()
	if r.decorateFn != nil {
		data = r.decorateFn(data)
	}
	r.masterPage().Update(data)
}

func (r *Resource) namespaceActions(aa ui.KeyActions) {
	if !r.list.Access(resource.NamespaceAccess) {
		return
	}
	r.namespaces = make(map[int]string, config.MaxFavoritesNS)
	// User can't list namespace. Don't offer a choice.
	if r.app.Conn() == nil || r.app.Conn().CheckListNSAccess() != nil {
		return
	}
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(resource.AllNamespace, r.switchNamespaceCmd, true)
	r.namespaces[0] = resource.AllNamespace
	index := 1
	for _, n := range r.app.Config.FavNamespaces() {
		if n == resource.AllNamespace {
			continue
		}
		aa[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(n, r.switchNamespaceCmd, true)
		r.namespaces[index] = n
		index++
	}
}

func (r *Resource) refreshActions() {
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", r.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Enter", r.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", r.refreshCmd, false),
	}
	aa[ui.KeySpace] = ui.NewKeyAction("Mark", r.markCmd, true)
	r.namespaceActions(aa)
	r.defaultActions(aa)

	if r.list.Access(resource.EditAccess) {
		aa[ui.KeyE] = ui.NewKeyAction("Edit", r.editCmd, true)
	}
	if r.list.Access(resource.DeleteAccess) {
		aa[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", r.deleteCmd, true)
	}
	if r.list.Access(resource.ViewAccess) {
		aa[ui.KeyY] = ui.NewKeyAction("YAML", r.viewCmd, true)
	}
	if r.list.Access(resource.DescribeAccess) {
		aa[ui.KeyD] = ui.NewKeyAction("Describe", r.describeCmd, true)
	}
	r.customActions(aa)

	t := r.masterPage()
	t.AddActions(aa)
}

func (r *Resource) customActions(aa ui.KeyActions) {
	pp := config.NewPlugins()
	if err := pp.Load(); err != nil {
		log.Warn().Msgf("No plugin configuration found")
		return
	}

	for k, plugin := range pp.Plugin {
		if !in(plugin.Scopes, r.list.GetName()) {
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
			r.execCmd(plugin.Command, plugin.Background, plugin.Args...),
			true)
	}
}

func (r *Resource) execCmd(bin string, bg bool, args ...string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if !r.masterPage().RowSelected() {
			return evt
		}

		var (
			env = r.envFn()
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

		if run(true, r.app, bin, bg, aa...) {
			r.app.Flash().Info("Custom CMD launched!")
		} else {
			r.app.Flash().Info("Custom CMD failed!")
		}
		return nil
	}
}

func (r *Resource) defaultK9sEnv() K9sEnv {
	ns, n := namespaced(r.masterPage().GetSelectedItem())
	ctx, err := r.app.Conn().Config().CurrentContextName()
	if err != nil {
		ctx = "n/a"
	}
	cluster, err := r.app.Conn().Config().CurrentClusterName()
	if err != nil {
		cluster = "n/a"
	}
	user, err := r.app.Conn().Config().CurrentUserName()
	if err != nil {
		user = "n/a"
	}
	groups, err := r.app.Conn().Config().CurrentGroupNames()
	if err != nil {
		groups = []string{"n/a"}
	}
	var cfg string
	kcfg := r.app.Conn().Config().Flags().KubeConfig
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

	row := r.masterPage().GetRow()
	for i, r := range row {
		env["COL"+strconv.Itoa(i)] = r
	}

	return env
}

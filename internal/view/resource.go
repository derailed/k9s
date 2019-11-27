package view

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Resource represents a generic resource viewer.
type Resource struct {
	*Table

	namespaces map[int]string
	list       resource.List
	path       string
	gvr        string
	envFn      EnvFunc
	currentNS  string
}

// NewResource returns a new viewer.
func NewResource(title, gvr string, list resource.List) ResourceViewer {
	return &Resource{
		Table: NewTable(title),
		list:  list,
		gvr:   gvr,
	}
}

// Init watches all running pods in given namespace
func (r *Resource) Init(ctx context.Context) error {
	log.Debug().Msgf(">>> RESOURCE INIT %s", r.list.GetName())

	if err := r.Table.Init(ctx); err != nil {
		return err
	}
	r.envFn = r.defaultK9sEnv
	r.Table.setFilterFn(r.filterResource)
	r.setNamespace(r.App().Config.ActiveNamespace())
	r.refresh()
	row, _ := r.GetSelection()
	if row == 0 && r.GetRowCount() > 0 {
		r.Select(1, 0)
	}

	return nil
}

// SetPath sets parent selector.
func (r *Resource) SetPath(p string) {
	r.path = p
}

// GetTable returns the underlying table view.
func (r *Resource) GetTable() *Table { return r.Table }

// SetEnvFn sets the function to pull current viewer env vars.
func (r *Resource) SetEnvFn(f EnvFunc) {
	r.envFn = f
}

// Start initializes updates.
func (r *Resource) Start() {
	r.Stop()

	log.Debug().Msgf(">>>>>>> START %s", r.list.GetName())
	r.Table.Start()

	var ctx context.Context
	ctx, r.cancelFn = context.WithCancel(context.Background())
	r.update(ctx)
}

// Name returns the component name.
func (r *Resource) Name() string {
	return r.list.GetName()
}

func (r *Resource) List() resource.List {
	return r.list
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

// ----------------------------------------------------------------------------
// Actions()...

func (r *Resource) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.RowSelected() {
		return evt
	}

	_, n := namespaced(r.GetSelectedItem())
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	r.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		r.app.Flash().Err(err)
	}

	return nil
}

func (r *Resource) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("RES ENTER CMD...")
	// If in command mode run filter otherwise enter function.
	if r.filterCmd(evt) == nil || !r.RowSelected() {
		return nil
	}

	f := r.defaultEnter
	if r.enterFn != nil {
		log.Debug().Msgf("Found custom enter")
		f = r.enterFn
	}
	f(r.app, r.list.GetNamespace(), r.list.GetName(), r.GetSelectedItem())

	return nil
}

func (r *Resource) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	r.app.Flash().Info("Refreshing...")
	r.refresh()
	return nil
}

func (r *Resource) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.RowSelected() {
		return evt
	}

	sel := r.GetSelectedItems()
	var msg string
	if len(sel) > 1 {
		msg = fmt.Sprintf("Delete %d marked %s?", len(sel), r.list.GetName())
	} else {
		msg = fmt.Sprintf("Delete %s %s?", r.list.GetName(), sel[0])
	}
	dialog.ShowDelete(r.app.Content.Pages, msg, func(cascade, force bool) {
		r.ShowDeleted()
		if len(sel) > 1 {
			r.app.Flash().Infof("Delete %d marked %s", len(sel), r.list.GetName())
		} else {
			r.app.Flash().Infof("Delete resource %s %s", r.list.GetName(), sel[0])
		}
		for _, res := range sel {
			if err := r.list.Resource().Delete(res, cascade, force); err != nil {
				r.app.Flash().Errf("Delete failed with %s", err)
			} else {
				r.app.forwarders.Kill(res)
			}
		}
		r.refresh()
	}, func() {})
	return nil
}

func (r *Resource) defaultEnter(app *App, ns, _, sel string) {
	if !r.list.Access(resource.DescribeAccess) {
		return
	}

	yaml, err := r.list.Resource().Describe(r.gvr, sel)
	if err != nil {
		r.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails("Describe")
	details.SetSubject(sel)
	details.SetTextColor(r.app.Styles.FgColor())
	details.SetText(colorizeYAML(r.app.Styles.Views().Yaml, yaml))
	details.ScrollToBeginning()
	r.app.inject(details)
}

func (r *Resource) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.RowSelected() {
		return evt
	}
	r.defaultEnter(r.app, r.list.GetNamespace(), r.list.GetName(), r.GetSelectedItem())

	return nil
}

func (r *Resource) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.RowSelected() {
		return evt
	}

	sel := r.GetSelectedItem()
	raw, err := r.list.Resource().Marshal(sel)
	if err != nil {
		r.app.Flash().Errf("Unable to marshal resource %s", err)
		return evt
	}
	details := NewDetails("YAML")
	details.SetSubject(sel)
	details.SetTextColor(r.app.Styles.FgColor())
	details.SetText(colorizeYAML(r.app.Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	r.app.inject(details)

	return nil
}

func (r *Resource) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.RowSelected() {
		return evt
	}

	r.Stop()
	{
		ns, po := namespaced(r.GetSelectedItem())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, r.list.GetName())
		args = append(args, "-n", ns)
		args = append(args, "--context", r.app.Config.K9s.CurrentContext)
		if cfg := r.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		if !runK(true, r.app, append(args, po)...) {
			r.app.Flash().Err(errors.New("Edit exec failed"))
		}
	}
	r.Start()

	return evt
}

func (r *Resource) setNamespace(ns string) {
	log.Debug().Msgf("!!!!!! SETTING NS %q", ns)
	if r.list.Namespaced() {
		r.currentNS = ns
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
	r.UpdateTitle()
	r.SelectRow(1, true)
	r.app.CmdBuff().Reset()
	if err := r.app.Config.SetActiveNamespace(r.currentNS); err != nil {
		log.Error().Err(err).Msg("Config save NS failed!")
	}
	if err := r.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}

	return nil
}

func (r *Resource) refresh() {
	log.Debug().Msgf("----> Refreshing (%q) -- %q -- `%s", r.currentNS, r.list.GetNamespace(), r.list.GetName())
	if r.list.Namespaced() {
		r.list.SetNamespace(r.currentNS)
	}

	if r.app.Conn() != nil {
		ctx := context.WithValue(context.Background(), resource.KeyFactory, r.app.factory)
		if err := r.list.Reconcile(ctx, r.gvr, r.path); err != nil {
			r.app.Flash().Err(err)
		}
	}
	data := r.list.Data()
	if r.decorateFn != nil {
		data = r.decorateFn(data)
	}
	r.refreshActions()
	r.Update(data)
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
	r.namespaceActions(aa)

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
	r.Actions().Set(aa)
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
		if !r.RowSelected() {
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
	return defaultK9sEnv(r.app, r.GetSelectedItem(), r.GetRow())
}

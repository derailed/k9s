package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

// ContextFunc enhances a given context.
type ContextFunc func(context.Context) context.Context

// Generic represents a generic resource vieweg.
type Generic struct {
	*Table

	namespaces map[int]string
	path       string
	gvr        dao.GVR
	envFn      EnvFunc
	meta       metav1.APIResource
	accessor   dao.Accessor
	contextFn  ContextFunc
}

// NewGeneric returns a new vieweg.
func NewGeneric(gvr dao.GVR) *Generic {
	return &Generic{
		Table: NewTable(string(gvr)),
		gvr:   gvr,
	}
}

// Init watches all running pods in given namespace
func (g *Generic) Init(ctx context.Context) error {
	log.Debug().Msgf(">>> GENERIC VIEW INIT %s", g.gvr)
	var err error
	g.meta, err = dao.MetaFor(g.gvr)
	if err != nil {
		return err
	}

	if err := g.Table.Init(ctx); err != nil {
		return err
	}
	g.Table.BaseTitle = g.meta.Kind
	g.accessor, err = dao.AccessorFor(g.app.factory, g.gvr)
	if err != nil {
		return err
	}

	g.envFn = g.defaultK9sEnv
	g.Table.setFilterFn(g.filterGeneric)
	g.setNamespace(g.App().Config.ActiveNamespace())
	g.refresh()
	row, _ := g.GetSelection()
	if row == 0 && g.GetRowCount() > 0 {
		g.Select(1, 0)
	}

	return nil
}

// Start initializes updates.
func (g *Generic) Start() {
	g.Stop()

	log.Debug().Msgf(">>>>>>> START %s", g.gvr)
	g.Table.Start()

	var ctx context.Context
	ctx, g.cancelFn = context.WithCancel(context.Background())
	go g.update(ctx)
}

func (g *Generic) Refresh() {
	g.app.QueueUpdateDraw(func() {
		g.refresh()
	})
}

// Name returns the component name.
func (g *Generic) Name() string {
	return g.meta.Kind
}

func (g *Generic) SetContextFn(f ContextFunc) {
	g.contextFn = f
}

// List returns a resource List.
func (g *Generic) List() resource.List { return nil }

// SetEnvFn sets a function to pull viewer env vars for plugins.
func (g *Generic) SetEnvFn(f EnvFunc) { g.envFn = f }

// SetPath set parents selector.
func (g *Generic) SetPath(p string) { g.Path = p }

// GVR returns a resource descriptor.
func (g *Generic) GVR() string { return string(g.gvr) }

func (g *Generic) GetTable() *Table {
	return g.Table
}
func (g *Generic) filterGeneric(sel string) {
	panic("NYI")
	// g.list.SetLabelSelector(sel)
	g.refresh()
}

func (g *Generic) update(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("%s updater canceled!", g.gvr)
			return
		case <-time.After(time.Duration(g.app.Config.K9s.GetRefreshRate()) * time.Second):
			g.app.QueueUpdateDraw(func() {
				g.refresh()
			})
		}
	}
}

// ----------------------------------------------------------------------------
// Actions()...

func (g *Generic) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !g.RowSelected() {
		return evt
	}

	_, n := namespaced(g.GetSelectedItem())
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	g.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		g.app.Flash().Err(err)
	}

	return nil
}

func (g *Generic) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("RES ENTER CMD...")
	// If in command mode run filter otherwise enter function.
	if g.filterCmd(evt) == nil || !g.RowSelected() {
		return nil
	}

	f := g.defaultEnter
	if g.enterFn != nil {
		log.Debug().Msgf("Found custom enter")
		f = g.enterFn
	}
	f(g.app, g.Data.Namespace, string(g.gvr), g.GetSelectedItem())

	return nil
}

func (g *Generic) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	g.app.Flash().Info("Refreshing...")
	g.refresh()
	return nil
}

func (g *Generic) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	selections := g.GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}
	log.Debug().Msgf("DEL SELECTIONS %#v", selections)

	var msg string
	if len(selections) > 1 {
		msg = fmt.Sprintf("Delete %d marked %s?", len(selections), g.gvr)
	} else {
		msg = fmt.Sprintf("Delete %s %s?", g.gvr, selections[0])
	}

	cancelFn := func() {}
	if in(g.meta.Categories, "K9s") {
		dialog.ShowConfirm(g.app.Content.Pages, "Confirm Delete", msg, func() {
			g.ShowDeleted()
			if len(selections) > 1 {
				g.app.Flash().Infof("Delete %d marked %s", len(selections), g.gvr)
			} else {
				g.app.Flash().Infof("Delete resource %s %s", g.gvr, selections[0])
			}
			for _, sel := range selections {
				ns, n := namespaced(sel)
				if err := g.accessor.(dao.Nuker).Delete(ns, n, true, true); err != nil {
					g.app.Flash().Errf("Delete failed with %s", err)
				} else {
					g.GetTable().DeleteMark(sel)
				}
			}
			g.refresh()
			g.SelectRow(1, true)
		}, cancelFn)
		return nil
	}

	dialog.ShowDelete(g.app.Content.Pages, msg, func(cascade, force bool) {
		g.ShowDeleted()
		if len(selections) > 1 {
			g.app.Flash().Infof("Delete %d marked %s", len(selections), g.gvr)
		} else {
			g.app.Flash().Infof("Delete resource %s %s", g.gvr, selections[0])
		}
		for _, sel := range selections {
			ns, n := namespaced(sel)
			if err := g.accessor.(dao.Nuker).Delete(ns, n, cascade, force); err != nil {
				g.app.Flash().Errf("Delete failed with %s", err)
			} else {
				g.app.forwarders.Kill(sel)
				g.GetTable().DeleteMark(sel)
			}
		}
		g.refresh()
		g.SelectRow(1, true)
	}, func() {})
	return nil
}

func (g *Generic) defaultEnter(app *App, ns, _, sel string) {
	log.Debug().Msgf("--------- Resource %q Verbs %v", sel, g.meta.Verbs)
	ns, n := namespaced(sel)
	yaml, err := dao.Describe(g.app.Conn(), g.gvr, ns, n)
	if err != nil {
		g.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails("Describe")
	details.SetSubject(sel)
	details.SetTextColor(g.app.Styles.FgColor())
	details.SetText(colorizeYAML(g.app.Styles.Views().Yaml, yaml))
	details.ScrollToBeginning()
	g.app.inject(details)
}

func (g *Generic) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("DESCRIBE %t -- %#v", g.RowSelected(), g.GetSelectedItems())
	if !g.RowSelected() {
		return evt
	}
	g.defaultEnter(g.app, g.Data.Namespace, string(g.gvr), g.GetSelectedItem())

	return nil
}

func (g *Generic) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !g.RowSelected() {
		return evt
	}

	sel := g.GetSelectedItem()
	ns, n := resource.Namespaced(sel)
	if ns == "" {
		ns = g.Data.Namespace
	}
	log.Debug().Msgf("------ NAMESPACES %q vs %q", ns, g.Data.Namespace)
	o, err := g.app.factory.Get(ns, string(g.gvr), n, labels.Everything())
	if err != nil {
		g.app.Flash().Errf("Unable to get resource %s", err)
		return nil
	}

	raw, err := toYAML(o)
	if err != nil {
		g.app.Flash().Errf("Unable to marshal resource %s", err)
		return nil
	}

	details := NewDetails("YAML")
	details.SetSubject(sel)
	details.SetTextColor(g.app.Styles.FgColor())
	details.SetText(colorizeYAML(g.app.Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	g.app.inject(details)

	return nil
}

func toYAML(o runtime.Object) (string, error) {
	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	err := p.PrintObj(o, &buff)
	if err != nil {
		log.Error().Msgf("Marshal Error %v", err)
		return "", err
	}

	return buff.String(), nil
}

func (g *Generic) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !g.RowSelected() {
		return evt
	}

	g.Stop()
	defer g.Start()
	{
		ns, po := namespaced(g.GetSelectedItem())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, g.meta.Kind)
		args = append(args, "-n", ns)
		args = append(args, "--context", g.app.Config.K9s.CurrentContext)
		if cfg := g.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		if !runK(true, g.app, append(args, po)...) {
			g.app.Flash().Err(errors.New("Edit exec failed"))
		}
	}

	return evt
}

func (g *Generic) setNamespace(ns string) {
	if !g.meta.Namespaced {
		g.Data.Namespace = render.ClusterWide
		return
	}
	if g.Data.Namespace == ns {
		return
	}

	if ns == render.NamespaceAll {
		ns = render.AllNamespaces
	}
	log.Debug().Msgf("!!!!!! SETTING NS %q", ns)
	g.Data.Namespace = ns
	g.Data.RowEvents = g.Data.RowEvents.Clear()
}

func (g *Generic) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := g.namespaces[i]
	if ns == "" {
		ns = render.NamespaceAll
	}

	g.app.switchNS(ns)
	g.setNamespace(ns)
	g.app.Flash().Infof("Viewing namespace `%s`...", ns)
	g.refresh()
	g.UpdateTitle()
	g.SelectRow(1, true)
	g.app.CmdBuff().Reset()
	if err := g.app.Config.SetActiveNamespace(g.Data.Namespace); err != nil {
		log.Error().Err(err).Msg("Config save NS failed!")
	}
	if err := g.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}

	return nil
}

func (g *Generic) refresh() {
	if g.app.Conn() == nil {
		log.Error().Msg("No api connection")
		return
	}

	log.Debug().Msgf("REFRESHING (%q) in ns %q", g.gvr, g.Data.Namespace)
	ctx := g.defaultContext()
	if g.contextFn != nil {
		ctx = g.contextFn(ctx)
	}
	data, err := dao.Reconcile(ctx, g.Table.Data, g.gvr)
	if err != nil {
		g.app.Flash().Err(err)
	}
	g.refreshActions()
	g.Update(data)
}

func (g *Generic) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, g.app.factory)
	ctx = context.WithValue(ctx, internal.KeySelection, g.Path)
	ctx = context.WithValue(ctx, internal.KeyLabels, "")
	ctx = context.WithValue(ctx, internal.KeyFields, "")

	return ctx
}

func (g *Generic) namespaceActions(aa ui.KeyActions) {
	if g.app.Conn() == nil || !g.meta.Namespaced {
		return
	}
	g.namespaces = make(map[int]string, config.MaxFavoritesNS)
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(resource.AllNamespace, g.switchNamespaceCmd, true)
	g.namespaces[0] = resource.AllNamespace
	index := 1
	for _, n := range g.app.Config.FavNamespaces() {
		if n == resource.AllNamespace {
			continue
		}
		aa[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(n, g.switchNamespaceCmd, true)
		g.namespaces[index] = n
		index++
	}
}

func (g *Generic) refreshActions() {
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", g.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", g.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", g.refreshCmd, false),
	}
	g.namespaceActions(aa)

	if dao.Can(g.meta.Verbs, "edit") {
		aa[ui.KeyE] = ui.NewKeyAction("Edit", g.editCmd, true)
	}
	if dao.Can(g.meta.Verbs, "delete") {
		aa[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", g.deleteCmd, true)
	}
	if dao.Can(g.meta.Verbs, "view") {
		aa[ui.KeyY] = ui.NewKeyAction("YAML", g.viewCmd, true)
	}
	if dao.Can(g.meta.Verbs, "describe") {
		aa[ui.KeyD] = ui.NewKeyAction("Describe", g.describeCmd, true)
	}
	g.customActions(aa)
	g.Actions().Set(aa)
}

func (g *Generic) customActions(aa ui.KeyActions) {
	pp := config.NewPlugins()
	if err := pp.Load(); err != nil {
		log.Warn().Msgf("No plugin configuration found")
		return
	}

	for k, plugin := range pp.Plugin {
		if !in(plugin.Scopes, g.meta.Name) {
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
			g.execCmd(plugin.Command, plugin.Background, plugin.Args...),
			true)
	}
}

func (g *Generic) execCmd(bin string, bg bool, args ...string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if !g.RowSelected() {

			return evt
		}

		var (
			env = g.envFn()
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

		if run(true, g.app, bin, bg, aa...) {
			g.app.Flash().Info("Custom CMD launched!")
		} else {
			g.app.Flash().Info("Custom CMD failed!")
		}
		return nil
	}
}

func (g *Generic) defaultK9sEnv() K9sEnv {
	return defaultK9sEnv(g.app, g.GetSelectedItem(), g.GetRow())
}

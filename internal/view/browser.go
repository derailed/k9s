package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	rt "runtime"
	"strconv"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
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

type BindKeysFunc func(ui.KeyActions)

// Browser represents a generic resource browser.
type Browser struct {
	*Table

	namespaces map[int]string
	gvr        client.GVR
	envFn      EnvFunc
	meta       metav1.APIResource
	accessor   dao.Accessor
	contextFn  ContextFunc
	bindKeysFn BindKeysFunc
	cancelFn   context.CancelFunc
}

// NewBrowser returns a new browser.
func NewBrowser(gvr client.GVR) ResourceViewer {
	return &Browser{
		Table: NewTable(string(gvr)),
		gvr:   gvr,
	}
}

// Init watches all running pods in given namespace
func (b *Browser) Init(ctx context.Context) error {
	log.Debug().Msgf("BROWSER INIT %s", b.gvr)
	var err error
	b.meta, err = dao.MetaFor(b.gvr)
	if err != nil {
		return err
	}

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	if !dao.IsK9sMeta(b.meta) {
		_ = b.app.factory.ForResource(b.app.Config.ActiveNamespace(), b.GVR())
		b.app.factory.WaitForCacheSync()
	}

	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.Table.BaseTitle = b.meta.Kind
	b.accessor, err = dao.AccessorFor(b.app.factory, b.gvr)
	if err != nil {
		return err
	}
	log.Debug().Msgf("ACCESSOR FOR %s -- %#v", b.gvr, b.accessor)

	b.envFn = b.defaultK9sEnv
	b.setNamespace(b.App().Config.ActiveNamespace())
	b.refresh()
	row, _ := b.GetSelection()
	if row == 0 && b.GetRowCount() > 0 {
		b.Select(1, 0)
	}

	return nil
}

// Start initializes updates.
func (b *Browser) Start() {
	b.Stop()

	log.Debug().Msgf("GOROUTINE %d", rt.NumGoroutine())

	log.Debug().Msgf("BROWSER START %s", b.gvr)
	b.Table.Start()

	var ctx context.Context
	ctx, b.cancelFn = context.WithCancel(context.Background())
	go b.update(ctx)
}

func (b *Browser) Stop() {
	if b.cancelFn != nil {
		b.cancelFn()
		b.cancelFn = nil
		log.Debug().Msgf("BROWSER <STOP> %s", b.BaseTitle)
	}
}

func (b *Browser) update(ctx context.Context) {
	defer log.Debug().Msgf("UPDATER BAIL For %s", b.gvr)
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("BROWSER <<CANCELED>> -- %s", b.gvr)
			return
		case <-time.After(time.Duration(b.app.Config.K9s.GetRefreshRate()) * time.Second):
			log.Debug().Msgf("GOROUTINE %d", rt.NumGoroutine())
			b.refresh()
		}
	}
}

// Name returns the component name.
func (b *Browser) Name() string { return b.meta.Kind }

// SetContextFn populates a custom context.
func (b *Browser) SetContextFn(f ContextFunc) { b.contextFn = f }

// SetBindKeysFn adds additional key bindings.
func (b *Browser) SetBindKeysFn(f BindKeysFunc) { b.bindKeysFn = f }

// SetEnvFn sets a function to pull viewer env vars for plugins.
func (b *Browser) SetEnvFn(f EnvFunc) { b.envFn = f }

// GVR returns a resource descriptor.
func (b *Browser) GVR() string { return string(b.gvr) }

// GetTable returns the underlying table.
func (b *Browser) GetTable() *Table { return b.Table }

// ----------------------------------------------------------------------------
// Actions()...

func (b *Browser) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.RowSelected() {
		return evt
	}

	_, n := client.Namespaced(b.GetSelectedItem())
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	b.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		b.app.Flash().Err(err)
	}

	return nil
}

func (b *Browser) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if b.filterCmd(evt) == nil || !b.RowSelected() {
		return nil
	}

	f := b.describeResource
	if b.enterFn != nil {
		f = b.enterFn
	}
	f(b.app, b.Data.Namespace, string(b.gvr), b.GetSelectedItem())

	return nil
}

func (b *Browser) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	b.app.Flash().Info("Refreshing...")
	b.refresh()
	return nil
}

func (b *Browser) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	selections := b.GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}
	log.Debug().Msgf("DEL SELECTIONS %#v", selections)

	b.Stop()
	defer b.Start()
	{
		msg := fmt.Sprintf("Delete %s %s?", b.gvr, selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), b.gvr)
		}

		cancelFn := func() {}
		if dao.IsK9sMeta(b.meta) {
			dialog.ShowConfirm(b.app.Content.Pages, "Confirm Delete", msg, func() {
				b.ShowDeleted()
				if len(selections) > 1 {
					b.app.Flash().Infof("Delete %d marked %s", len(selections), b.gvr)
				} else {
					b.app.Flash().Infof("Delete resource %s %s", b.gvr, selections[0])
				}
				for _, sel := range selections {
					if err := b.accessor.(dao.Nuker).Delete(sel, true, true); err != nil {
						b.app.Flash().Errf("Delete failed with `%s", err)
					} else {
						b.GetTable().DeleteMark(sel)
					}
				}
				b.refresh()
				b.SelectRow(1, true)
			}, cancelFn)
			return nil
		}

		dialog.ShowDelete(b.app.Content.Pages, msg, func(cascade, force bool) {
			b.ShowDeleted()
			if len(selections) > 1 {
				b.app.Flash().Infof("Delete %d marked %s", len(selections), b.gvr)
			} else {
				b.app.Flash().Infof("Delete resource %s %s", b.gvr, selections[0])
			}
			for _, sel := range selections {
				if err := b.accessor.(dao.Nuker).Delete(sel, cascade, force); err != nil {
					b.app.Flash().Errf("Delete failed with `%s", err)
				} else {
					b.app.factory.DeleteForwarder(sel)
					b.GetTable().DeleteMark(sel)
				}
			}
			b.refresh()
			b.SelectRow(1, true)
		}, cancelFn)
	}

	return nil
}

func (b *Browser) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("DESCRIBE %t -- %#v", b.RowSelected(), b.GetSelectedItems())
	if !b.RowSelected() {
		return evt
	}
	b.describeResource(b.app, b.Data.Namespace, string(b.gvr), b.GetSelectedItem())

	return nil
}

func (b *Browser) describeResource(app *App, _, _, sel string) {
	ns, n := client.Namespaced(sel)
	yaml, err := dao.Describe(b.app.Conn(), b.gvr, ns, n)
	if err != nil {
		b.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails("Describe")
	details.SetSubject(sel)
	details.SetTextColor(b.app.Styles.FgColor())
	details.SetText(colorizeYAML(b.app.Styles.Views().Yaml, yaml))
	details.ScrollToBeginning()
	if err := b.app.inject(details); err != nil {
		b.app.Flash().Err(err)
	}
}

func (b *Browser) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.RowSelected() {
		return evt
	}

	path := b.GetSelectedItem()
	log.Debug().Msgf("------ NAMESPACES %q vs %q", path, b.Data.Namespace)
	o, err := b.app.factory.Get(string(b.gvr), path, labels.Everything())
	if err != nil {
		b.app.Flash().Errf("Unable to get resource %q -- %s", b.gvr, err)
		return nil
	}

	raw, err := toYAML(o)
	if err != nil {
		b.app.Flash().Errf("Unable to marshal resource %s", err)
		return nil
	}

	details := NewDetails("YAML")
	details.SetSubject(path)
	details.SetTextColor(b.app.Styles.FgColor())
	details.SetText(colorizeYAML(b.app.Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	if err := b.app.inject(details); err != nil {
		b.App().Flash().Err(err)
	}

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

func (b *Browser) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.RowSelected() {
		return evt
	}

	b.Stop()
	defer b.Start()
	{
		ns, po := client.Namespaced(b.GetSelectedItem())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, b.meta.Kind)
		args = append(args, "-n", ns)
		args = append(args, "--context", b.app.Config.K9s.CurrentContext)
		if cfg := b.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		if !runK(true, b.app, append(args, po)...) {
			b.app.Flash().Err(errors.New("Edit exec failed"))
		}
	}

	return evt
}

func (b *Browser) setNamespace(ns string) {
	if !b.meta.Namespaced {
		b.Data.Namespace = render.ClusterScope
		return
	}
	if b.Data.Namespace == ns {
		return
	}

	if ns == render.NamespaceAll {
		ns = render.AllNamespaces
	}
	log.Debug().Msgf("!!!!!! SETTING NS %q", ns)
	b.Data.Namespace = ns
	b.Data.RowEvents = b.Data.RowEvents.Clear()
}

func (b *Browser) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := b.namespaces[i]
	if ns == "" {
		ns = render.NamespaceAll
	}

	b.app.switchNS(ns)
	b.setNamespace(ns)
	b.app.Flash().Infof("Viewing namespace `%s`...", ns)
	b.refresh()
	b.UpdateTitle()
	b.SelectRow(1, true)
	b.app.CmdBuff().Reset()
	if err := b.app.Config.SetActiveNamespace(b.Data.Namespace); err != nil {
		log.Error().Err(err).Msg("Config save NS failed!")
	}
	if err := b.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}

	return nil
}

func (b *Browser) refresh() {
	if b.app.Conn() == nil {
		return
	}
	ctx := b.defaultContext()
	if b.contextFn != nil {
		ctx = b.contextFn(ctx)
	}
	if path, ok := ctx.Value(internal.KeyPath).(string); ok && path != "" {
		b.Path = path
	}
	data, err := dao.Reconcile(ctx, b.Table.Data, b.gvr)
	b.app.QueueUpdateDraw(func() {
		if err != nil {
			b.app.Flash().Err(err)
		}
		b.refreshActions()
		b.Update(data)
	})
}

func (b *Browser) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, b.app.factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, string(b.gvr))
	ctx = context.WithValue(ctx, internal.KeyPath, b.Path)
	ctx = context.WithValue(ctx, internal.KeyLabels, "")
	ctx = context.WithValue(ctx, internal.KeyFields, "")

	return ctx
}

func (b *Browser) namespaceActions(aa ui.KeyActions) {
	if b.app.Conn() == nil || !b.meta.Namespaced || b.GetTable().Path != "" {
		log.Warn().Msgf("NOT NAMESPACE RES %q -- %t -- %q", b.gvr, b.meta.Namespaced, b.GetTable().Path)
		return
	}
	b.namespaces = make(map[int]string, config.MaxFavoritesNS)
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(render.NamespaceAll, b.switchNamespaceCmd, true)
	b.namespaces[0] = render.NamespaceAll
	index := 1
	for _, n := range b.app.Config.FavNamespaces() {
		if n == render.NamespaceAll {
			continue
		}
		aa[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(n, b.switchNamespaceCmd, true)
		b.namespaces[index] = n
		index++
	}
}

func (b *Browser) refreshActions() {
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", b.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", b.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", b.refreshCmd, false),
	}
	b.namespaceActions(aa)

	if client.Can(b.meta.Verbs, "edit") {
		aa[ui.KeyE] = ui.NewKeyAction("Edit", b.editCmd, true)
	}
	if client.Can(b.meta.Verbs, "delete") {
		aa[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", b.deleteCmd, true)
	}
	if client.Can(b.meta.Verbs, "view") {
		aa[ui.KeyY] = ui.NewKeyAction("YAML", b.viewCmd, true)
	}
	if client.Can(b.meta.Verbs, "describe") {
		aa[ui.KeyD] = ui.NewKeyAction("Describe", b.describeCmd, true)
	}
	b.customActions(aa)
	b.Actions().Add(aa)

	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.app.Menu().HydrateMenu(b.Hints())
}

func (b *Browser) customActions(aa ui.KeyActions) {
	pp := config.NewPlugins()
	if err := pp.Load(); err != nil {
		log.Warn().Msgf("No plugin configuration found")
		return
	}

	for k, plugin := range pp.Plugin {
		if !in(plugin.Scopes, b.meta.Name) {
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
			b.execCmd(plugin.Command, plugin.Background, plugin.Args...),
			true)
	}
}

func (b *Browser) execCmd(bin string, bg bool, args ...string) ui.ActionHandler {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if !b.RowSelected() {

			return evt
		}

		var (
			env = b.envFn()
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

		if run(true, b.app, bin, bg, aa...) {
			b.app.Flash().Info("Custom CMD launched!")
		} else {
			b.app.Flash().Info("Custom CMD failed!")
		}
		return nil
	}
}

func (b *Browser) defaultK9sEnv() K9sEnv {
	return defaultK9sEnv(b.app, b.GetSelectedItem(), b.GetRow())
}

package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

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

// Browser represents a generic resource browser.
type Browser struct {
	*Table

	namespaces map[int]string
	gvr        client.GVR
	meta       metav1.APIResource
	accessor   dao.Accessor
	envFn      EnvFunc
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
	var err error
	b.meta, err = dao.MetaFor(b.gvr)
	if err != nil {
		return err
	}

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	if !dao.IsK9sMeta(b.meta) {
		if _, e := b.app.factory.CanForResource(b.app.Config.ActiveNamespace(), b.GVR()); e != nil {
			return e
		}
	}

	b.bindKeys()
	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.BaseTitle = b.meta.Kind
	b.accessor, err = dao.AccessorFor(b.app.factory, b.gvr)
	if err != nil {
		return err
	}

	b.envFn = b.defaultK9sEnv
	b.setNamespace(b.App().Config.ActiveNamespace())
	row, _ := b.GetSelection()
	if row == 0 && b.GetRowCount() > 0 {
		b.Select(1, 0)
	}
	b.GetModel().AddListener(b)
	b.App().Status(ui.FlashWarn, "Loading...")

	return nil
}

func (b *Browser) bindKeys() {
	b.Actions().Add(ui.KeyActions{
		tcell.KeyEscape: ui.NewSharedKeyAction("Filter Reset", b.resetCmd, false),
		tcell.KeyEnter:  ui.NewSharedKeyAction("Filter", b.filterCmd, false),
	})
}

// Start initializes browser updates.
func (b *Browser) Start() {
	b.Stop()

	b.Table.Start()
	ctx := b.defaultContext()
	ctx, b.cancelFn = context.WithCancel(ctx)
	if b.contextFn != nil {
		ctx = b.contextFn(ctx)
	}
	if path, ok := ctx.Value(internal.KeyPath).(string); ok && path != "" {
		b.Path = path
	}
	b.GetModel().Watch(ctx)
}

func (b *Browser) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.SearchBuff().InCmdMode() {
		b.SearchBuff().Reset()
		return b.App().PrevCmd(evt)
	}

	cmd := b.SearchBuff().String()
	b.App().Flash().Info("Clearing filter...")
	b.SearchBuff().Reset()

	if ui.IsLabelSelector(cmd) {
		b.Start()
	} else {
		b.Refresh()
	}

	return nil
}

func (b *Browser) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.SearchBuff().IsActive() {
		return evt
	}

	b.SearchBuff().SetActive(false)

	cmd := b.SearchBuff().String()
	if ui.IsLabelSelector(cmd) {
		b.Start()
		return nil
	}
	b.Refresh()

	return nil
}

// Stop terminates browser updates.
func (b *Browser) Stop() {
	if b.cancelFn == nil {
		return
	}
	b.Table.Stop()
	log.Debug().Msgf("BROWSER <STOP> %q", b.gvr)
	b.cancelFn()
	b.cancelFn = nil
}

func (b *Browser) refresh() {
	b.Start()
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

// TableLoadFailed notifies view something went south.
func (b *Browser) TableLoadFailed(err error) {
	b.app.QueueUpdateDraw(func() {
		b.app.Flash().Err(err)
		b.App().ClearStatus()
	})
}

// TableDataChanged notifies view new data is available.
func (b *Browser) TableDataChanged(data render.TableData) {
	b.app.QueueUpdateDraw(func() {
		b.refreshActions()
		b.Update(data)
		b.App().ClearStatus()
	})
}

// ----------------------------------------------------------------------------
// Actions()...

func (b *Browser) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}

	_, n := client.Namespaced(path)
	log.Debug().Msgf("Copied selection to clipboard %q", n)
	b.app.Flash().Info("Current selection copied to clipboard...")
	if err := clipboard.WriteAll(n); err != nil {
		b.app.Flash().Err(err)
	}

	return nil
}

func (b *Browser) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if b.filterCmd(evt) == nil || path == "" {
		return nil
	}

	f := b.describeResource
	if b.enterFn != nil {
		f = b.enterFn
	}
	f(b.app, b.GetModel().GetNamespace(), string(b.gvr), path)

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

	b.Stop()
	defer b.Start()
	{
		msg := fmt.Sprintf("Delete %s %s?", b.gvr, selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), b.gvr)
		}
		if dao.IsK9sMeta(b.meta) {
			b.simpleDelete(selections, msg)
			return nil
		}
		b.resourceDelete(selections, msg)
	}

	return nil
}

func (b *Browser) simpleDelete(selections []string, msg string) {
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
	}, func() {})
}

func (b *Browser) resourceDelete(selections []string, msg string) {
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
	}, func() {})
}

func (b *Browser) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}
	b.describeResource(b.app, b.GetModel().GetNamespace(), string(b.gvr), path)

	return nil
}

func (b *Browser) describeResource(app *App, _, _, sel string) {
	ns, n := client.Namespaced(sel)
	yaml, err := dao.Describe(b.app.Conn(), b.gvr, ns, n)
	if err != nil {
		b.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails(b.App(), "Describe", sel).Update(yaml)
	if err := b.app.inject(details); err != nil {
		b.app.Flash().Err(err)
	}
}

func (b *Browser) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}

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

	details := NewDetails(b.app, "YAML", path).Update(raw)
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
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}

	b.Stop()
	defer b.Start()
	{
		ns, n := client.Namespaced(path)
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, b.meta.Kind)
		args = append(args, "-n", ns)
		args = append(args, "--context", b.app.Config.K9s.CurrentContext)
		if cfg := b.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		if !runK(true, b.app, append(args, n)...) {
			b.app.Flash().Err(errors.New("Edit exec failed"))
		}
	}

	return evt
}

func (b *Browser) setNamespace(ns string) {
	if !b.meta.Namespaced {
		b.GetModel().SetNamespace(render.ClusterScope)
		return
	}
	if b.GetModel().InNamespace(ns) {
		return
	}

	if ns == render.NamespaceAll {
		ns = render.AllNamespaces
	}
	b.GetModel().SetNamespace(ns)
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
	if err := b.app.Config.SetActiveNamespace(b.GetModel().GetNamespace()); err != nil {
		log.Error().Err(err).Msg("Config save NS failed!")
	}
	if err := b.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}

	return nil
}

func (b *Browser) defaultContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, internal.KeyFactory, b.app.factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, string(b.gvr))
	ctx = context.WithValue(ctx, internal.KeyPath, b.Path)

	ctx = context.WithValue(ctx, internal.KeyLabels, "")
	if ui.IsLabelSelector(b.SearchBuff().String()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(b.SearchBuff().String()))
	}
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyNamespace, b.App().Config.ActiveNamespace())

	return ctx
}

func (b *Browser) namespaceActions(aa ui.KeyActions) {
	if b.app.Conn() == nil || !b.meta.Namespaced || b.GetTable().Path != "" {
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
	pluginActions(b, aa)
	hotKeyActions(b, aa)
	b.Actions().Add(aa)

	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.app.Menu().HydrateMenu(b.Hints())
}

// Aliases returns all available aliases.
func (b *Browser) Aliases() []string {
	return append(b.meta.ShortNames, b.meta.SingularName, b.meta.Name)
}

// EnvFn returns an plugin env function if available.
func (b *Browser) EnvFn() EnvFunc {
	return b.envFn
}

func (b *Browser) defaultK9sEnv() K9sEnv {
	return defaultK9sEnv(b.app, b.GetSelectedItem(), b.GetSelectedRow())
}

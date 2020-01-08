package view

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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
)

// Browser represents a generic resource browser.
type Browser struct {
	*Table

	namespaces map[int]string
	meta       metav1.APIResource
	accessor   dao.Accessor
	contextFn  ContextFunc
	cancelFn   context.CancelFunc
}

// NewBrowser returns a new browser.
func NewBrowser(gvr client.GVR) ResourceViewer {
	return &Browser{
		Table: NewTable(gvr),
	}
}

// Init watches all running pods in given namespace
func (b *Browser) Init(ctx context.Context) error {
	var err error
	b.meta, err = dao.MetaFor(b.gvr)
	if err != nil {
		return err
	}
	b.BaseTitle = b.meta.Kind

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	ns := b.app.Config.ActiveNamespace()
	if dao.IsK8sMeta(b.meta) {
		if _, e := b.app.factory.CanForResource(ns, b.GVR(), client.MonitorAccess); e != nil {
			return e
		}
	}

	b.bindKeys()
	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.accessor, err = dao.AccessorFor(b.app.factory, b.gvr)
	if err != nil {
		return err
	}

	b.setNamespace(ns)
	row, _ := b.GetSelection()
	if row == 0 && b.GetRowCount() > 0 {
		b.Select(1, 0)
	}
	b.GetModel().AddListener(b)
	b.GetModel().SetRefreshRate(time.Duration(b.App().Config.K9s.GetRefreshRate()) * time.Second)

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

	b.App().Status(ui.FlashInfo, "Loading...")
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

// Stop terminates browser updates.
func (b *Browser) Stop() {
	if b.cancelFn == nil {
		return
	}
	b.Table.Stop()
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

// GetTable returns the underlying table.
func (b *Browser) GetTable() *Table { return b.Table }

// Aliases returns all available aliases.
func (b *Browser) Aliases() []string {
	return append(b.meta.ShortNames, b.meta.SingularName, b.meta.Name)
}

// ----------------------------------------------------------------------------
// Model Protocol...

// TableDataChanged notifies view new data is available.
func (b *Browser) TableDataChanged(data render.TableData) {
	b.app.QueueUpdateDraw(func() {
		b.refreshActions()
		b.Update(data)
		b.App().ClearStatus(true)
	})
}

// TableLoadFailed notifies view something went south.
func (b *Browser) TableLoadFailed(err error) {
	b.app.QueueUpdateDraw(func() {
		b.app.Flash().Err(err)
		b.App().ClearStatus(false)
	})
}

// ----------------------------------------------------------------------------
// Actions...

func (b *Browser) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}

	ctx := b.defaultContext()
	raw, err := b.GetModel().ToYAML(ctx, path)
	if err != nil {
		b.App().Flash().Errf("unable to get resource %q -- %s", b.gvr, err)
		return nil
	}

	details := NewDetails(b.app, "YAML", path).Update(raw)
	if err := b.App().inject(details); err != nil {
		b.App().Flash().Err(err)
	}

	return nil
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

func (b *Browser) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if b.filterCmd(evt) == nil || path == "" {
		return nil
	}

	f := describeResource
	if b.enterFn != nil {
		f = b.enterFn
	}
	f(b.app, b.GetModel(), b.gvr.String(), path)

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
		msg := fmt.Sprintf("Delete %s %s?", b.gvr.ToR(), selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), b.gvr)
		}
		if !dao.IsK8sMeta(b.meta) {
			b.simpleDelete(selections, msg)
			return nil
		}
		b.resourceDelete(selections, msg)
	}

	return nil
}

func (b *Browser) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}
	describeResource(b.app, b.GetModel(), b.gvr.String(), path)

	return nil
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

func (b *Browser) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, _ := strconv.Atoi(string(evt.Rune()))
	ns := b.namespaces[i]
	if ns == "" {
		ns = client.NamespaceAll
	}

	auth, err := b.App().factory.Client().CanI(ns, b.GVR(), client.MonitorAccess)
	if !auth {
		b.App().Flash().Err(err)
		return nil
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

// ----------------------------------------------------------------------------
// Helpers...

func (b *Browser) setNamespace(ns string) {
	if !b.meta.Namespaced {
		b.GetModel().SetNamespace(client.ClusterScope)
		return
	}
	if b.GetModel().InNamespace(ns) {
		return
	}

	b.GetModel().SetNamespace(client.NormalizeNS(ns))
}

func (b *Browser) defaultContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, internal.KeyFactory, b.app.factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, b.gvr.String())
	ctx = context.WithValue(ctx, internal.KeyPath, b.Path)

	ctx = context.WithValue(ctx, internal.KeyLabels, "")
	if ui.IsLabelSelector(b.SearchBuff().String()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(b.SearchBuff().String()))
	}
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.NormalizeNS(b.App().Config.ActiveNamespace()))

	return ctx
}

func (b *Browser) namespaceActions(aa ui.KeyActions) {
	if b.app.Conn() == nil || !b.meta.Namespaced || b.GetTable().Path != "" {
		return
	}
	b.namespaces = make(map[int]string, config.MaxFavoritesNS)
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(client.NamespaceAll, b.switchNamespaceCmd, true)
	b.namespaces[0] = client.NamespaceAll
	index := 1
	for _, n := range b.app.Config.FavNamespaces() {
		if n == client.NamespaceAll {
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

	if !dao.IsK9sMeta(b.meta) {
		aa[ui.KeyY] = ui.NewKeyAction("YAML", b.viewCmd, true)
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

func (b *Browser) simpleDelete(selections []string, msg string) {
	dialog.ShowConfirm(b.app.Content.Pages, "Confirm Delete", msg, func() {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.gvr)
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.gvr, selections[0])
		}
		for _, sel := range selections {
			log.Debug().Msgf("YO!! %#v", b.accessor)
			nuker, ok := b.accessor.(dao.Nuker)
			if !ok {
				b.app.Flash().Errf("Invalid nuker %T", b.accessor)
				return
			}
			if err := nuker.Delete(sel, true, true); err != nil {
				b.app.Flash().Errf("Delete failed with `%s", err)
			} else {
				b.GetTable().DeleteMark(sel)
			}
		}
		b.refresh()
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
			if err := b.GetModel().Delete(b.defaultContext(), sel, cascade, force); err != nil {
				b.app.Flash().Errf("Delete failed with `%s", err)
			} else {
				b.app.Flash().Infof("%s `%s deleted successfully", b.GVR(), sel)
				b.app.factory.DeleteForwarder(sel)
				b.GetTable().DeleteMark(sel)
			}
		}
		b.refresh()
	}, func() {})
}

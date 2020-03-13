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
	b.meta, err = dao.MetaAccess.MetaFor(b.GVR())
	if err != nil {
		return err
	}

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	ns := client.CleanseNamespace(b.app.Config.ActiveNamespace())
	if dao.IsK8sMeta(b.meta) && b.app.ConOK() {
		if _, e := b.app.factory.CanForResource(ns, b.GVR().String(), client.MonitorAccess); e != nil {
			return e
		}
	}
	b.app.CmdBuff().Reset()

	b.bindKeys()
	if b.bindKeysFn != nil {
		b.bindKeysFn(b.Actions())
	}
	b.accessor, err = dao.AccessorFor(b.app.factory, b.GVR())
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

// SetInstance sets a single instance view.
func (b *Browser) SetInstance(path string) {
	b.GetModel().SetInstance(path)
}

// Start initializes browser updates.
func (b *Browser) Start() {
	b.Stop()
	b.Table.Start()
	b.GetModel().Watch(b.prepareContext())
}

func (b *Browser) prepareContext() context.Context {
	ctx := b.defaultContext()
	ctx, b.cancelFn = context.WithCancel(ctx)
	if b.contextFn != nil {
		ctx = b.contextFn(ctx)
	}
	if path, ok := ctx.Value(internal.KeyPath).(string); ok && path != "" {
		b.Path = path
	}

	return ctx
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
	if !b.app.ConOK() {
		return
	}
	b.app.QueueUpdateDraw(func() {
		b.refreshActions()
		b.Update(data)
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
		b.App().Flash().Errf("unable to get resource %q -- %s", b.GVR(), err)
		return nil
	}

	details := NewDetails(b.app, "YAML", path, true).Update(raw)
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
	f(b.app, b.GetModel(), b.GVR().String(), path)

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
		msg := fmt.Sprintf("Delete %s %s?", b.GVR().R(), selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), b.GVR())
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
	describeResource(b.app, b.GetModel(), b.GVR().String(), path)

	return nil
}

func (b *Browser) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt

	}
	ns, n := client.Namespaced(path)

	if ok, err := b.app.Conn().CanI(ns, b.GVR().String(), []string{"patch"}); !ok || err != nil {
		b.App().Flash().Err(fmt.Errorf("Current user can't edit resource %s", b.GVR()))
		return nil
	}

	b.Stop()
	defer b.Start()
	{
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, b.meta.SingularName)
		args = append(args, "-n", ns)
		if !runK(b.app, shellOpts{clear: true, args: append(args, n)}) {
			b.app.Flash().Err(errors.New("Edit exec failed"))
		}
	}

	return evt
}

func (b *Browser) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, err := strconv.Atoi(string(evt.Rune()))
	if err != nil {
		log.Error().Err(err).Msgf("Fail to switch namespace")
		return nil
	}
	ns := b.namespaces[i]

	auth, err := b.App().factory.Client().CanI(ns, b.GVR().String(), client.MonitorAccess)
	if !auth {
		if err == nil {
			err = fmt.Errorf("current user can't access namespace %s", ns)
		}
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
	ns = client.CleanseNamespace(ns)
	if b.GetModel().InNamespace(ns) {
		return
	}
	if !b.meta.Namespaced {
		ns = client.ClusterScope
	}
	b.GetModel().SetNamespace(client.CleanseNamespace(ns))
}

func (b *Browser) defaultContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, internal.KeyFactory, b.app.factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, b.GVR().String())
	ctx = context.WithValue(ctx, internal.KeyPath, b.Path)

	ctx = context.WithValue(ctx, internal.KeyLabels, "")
	if ui.IsLabelSelector(b.SearchBuff().String()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(b.SearchBuff().String()))
	}
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.CleanseNamespace(b.App().Config.ActiveNamespace()))

	return ctx
}

func (b *Browser) refreshActions() {
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", b.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", b.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", b.refreshCmd, false),
	}

	if b.app.ConOK() {
		b.namespaceActions(aa)
		if !b.app.Config.K9s.GetReadOnly() {
			if client.Can(b.meta.Verbs, "edit") {
				aa[ui.KeyE] = ui.NewKeyAction("Edit", b.editCmd, true)
			}
			if client.Can(b.meta.Verbs, "delete") {
				aa[tcell.KeyCtrlD] = ui.NewKeyAction("Delete", b.deleteCmd, true)
			}
		}
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

func (b *Browser) namespaceActions(aa ui.KeyActions) {
	if !b.meta.Namespaced || b.GetTable().Path != "" {
		return
	}
	b.namespaces = make(map[int]string, config.MaxFavoritesNS)
	aa[tcell.Key(ui.NumKeys[0])] = ui.NewKeyAction(client.NamespaceAll, b.switchNamespaceCmd, true)
	b.namespaces[0] = client.NamespaceAll
	index := 1
	for _, ns := range b.app.Config.FavNamespaces() {
		if ns == client.NamespaceAll {
			continue
		}
		aa[tcell.Key(ui.NumKeys[index])] = ui.NewKeyAction(ns, b.switchNamespaceCmd, true)
		b.namespaces[index] = ns
		index++
	}
}

func (b *Browser) simpleDelete(selections []string, msg string) {
	dialog.ShowConfirm(b.app.Content.Pages, "Confirm Delete", msg, func() {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
		}
		for _, sel := range selections {
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
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
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

package view

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell/v2"
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
	mx         sync.RWMutex
}

// NewBrowser returns a new browser.
func NewBrowser(gvr client.GVR) ResourceViewer {
	return &Browser{
		Table: NewTable(gvr),
	}
}

// Init watches all running pods in given namespace.
func (b *Browser) Init(ctx context.Context) error {
	var err error
	b.meta, err = dao.MetaAccess.MetaFor(b.GVR())
	if err != nil {
		return err
	}
	colorerFn := render.DefaultColorer
	if r, ok := model.Registry[b.GVR().String()]; ok {
		colorerFn = r.Renderer.ColorerFunc()
	}
	b.GetTable().SetColorerFn(colorerFn)

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	ns := client.CleanseNamespace(b.app.Config.ActiveNamespace())
	if dao.IsK8sMeta(b.meta) && b.app.ConOK() {
		if _, e := b.app.factory.CanForResource(ns, b.GVR().String(), client.MonitorAccess); e != nil {
			return e
		}
	}
	if b.App().IsRunning() {
		b.app.CmdBuff().Reset()
	}

	b.bindKeys(b.Actions())
	for _, f := range b.bindKeysFn {
		f(b.Actions())
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
	b.GetModel().SetRefreshRate(time.Duration(b.App().Config.K9s.GetRefreshRate()) * time.Second)

	b.CmdBuff().SetSuggestionFn(b.suggestFilter())

	return nil
}

// InCmdMode checks if prompt is active.
func (b *Browser) InCmdMode() bool {
	return b.CmdBuff().InCmdMode()
}

func (b *Browser) suggestFilter() model.SuggestionFunc {
	return func(s string) (entries sort.StringSlice) {
		if s == "" {
			if b.App().filterHistory.Empty() {
				return
			}
			return b.App().filterHistory.List()
		}

		s = strings.ToLower(s)
		for _, h := range b.App().filterHistory.List() {
			if s == h {
				continue
			}
			if strings.HasPrefix(h, s) {
				entries = append(entries, strings.Replace(h, s, "", 1))
			}
		}
		return
	}
}

func (b *Browser) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyEscape: ui.NewSharedKeyAction("Filter Reset", b.resetCmd, false),
		tcell.KeyEnter:  ui.NewSharedKeyAction("Filter", b.filterCmd, false),
		tcell.KeyHelp:   ui.NewSharedKeyAction("Help", b.helpCmd, false),
	})
}

// SetInstance sets a single instance view.
func (b *Browser) SetInstance(path string) {
	b.GetModel().SetInstance(path)
}

// Start initializes browser updates.
func (b *Browser) Start() {
	b.app.Config.ValidateFavorites()
	ns := b.app.Config.ActiveNamespace()
	if n := b.GetModel().GetNamespace(); !client.IsClusterScoped(n) {
		ns = n
	}
	if err := b.app.switchNS(ns); err != nil {
		log.Error().Err(err).Msgf("ns switch failed")
	}

	b.Stop()
	b.GetModel().AddListener(b)
	b.Table.Start()
	b.CmdBuff().AddListener(b)
	if err := b.GetModel().Watch(b.prepareContext()); err != nil {
		b.App().Flash().Err(fmt.Errorf("Watcher failed for %s -- %w", b.GVR(), err))
	}
}

// Stop terminates browser updates.
func (b *Browser) Stop() {
	b.mx.Lock()
	{
		if b.cancelFn != nil {
			b.cancelFn()
			b.cancelFn = nil
		}
	}
	b.mx.Unlock()
	b.GetModel().RemoveListener(b)
	b.CmdBuff().RemoveListener(b)
	b.Table.Stop()
}

// BufferChanged indicates the buffer was changed.
func (b *Browser) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (b *Browser) BufferCompleted(text, _ string) {
	if ui.IsLabelSelector(text) {
		b.GetModel().SetLabelFilter(ui.TrimLabelSelector(text))
	} else {
		b.GetModel().SetLabelFilter("")
	}
}

// BufferActive indicates the buff activity changed.
func (b *Browser) BufferActive(state bool, k model.BufferKind) {
	if state {
		return
	}
	if err := b.GetModel().Refresh(b.prepareContext()); err != nil {
		log.Error().Err(err).Msgf("Refresh failed for %s", b.GVR())
	}
	b.app.QueueUpdateDraw(func() {
		b.Update(b.GetModel().Peek(), b.App().Conn().HasMetrics())
		if b.GetRowCount() > 1 {
			b.App().filterHistory.Push(b.CmdBuff().GetText())
		}
	})
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
func (b *Browser) TableDataChanged(data *render.TableData) {
	var cancel context.CancelFunc
	b.mx.RLock()
	cancel = b.cancelFn
	b.mx.RUnlock()

	if !b.app.ConOK() || cancel == nil || !b.app.IsRunning() {
		return
	}

	b.app.QueueUpdateDraw(func() {
		b.refreshActions()
		b.Update(data, b.app.Conn().HasMetrics())
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

	v := NewLiveView(b.app, "YAML", model.NewYAML(b.GVR(), path))
	if err := v.app.inject(v); err != nil {
		v.app.Flash().Err(err)
	}
	return nil
}

func (b *Browser) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if b.CmdBuff().InCmdMode() {
		return nil
	}

	return evt
}

func (b *Browser) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.CmdBuff().InCmdMode() {
		b.CmdBuff().ClearText(false)
		return b.App().PrevCmd(evt)
	}

	b.CmdBuff().Reset()
	if ui.IsLabelSelector(b.CmdBuff().GetText()) {
		b.Start()
	}
	b.Refresh()

	return nil
}

func (b *Browser) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !b.CmdBuff().IsActive() {
		return evt
	}

	b.CmdBuff().SetActive(false)
	if ui.IsLabelSelector(b.CmdBuff().GetText()) {
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
	if client.IsClusterScoped(ns) {
		ns = client.AllNamespaces
	}
	if b.GVR().String() == "v1/namespaces" {
		ns = n
	}
	if ok, err := b.app.Conn().CanI(ns, b.GVR().String(), []string{"patch"}); !ok || err != nil {
		b.App().Flash().Err(fmt.Errorf("Current user can't edit resource %s", b.GVR()))
		return nil
	}

	b.Stop()
	defer b.Start()
	{
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, b.GVR().FQN(n))
		if ns != client.AllNamespaces {
			args = append(args, "-n", ns)
		}
		if !runK(b.app, shellOpts{clear: true, args: args}) {
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

	if client.IsAllNamespace(ns) {
		b.GetTable().SetSortCol("NAMESPACE", true)
	}

	if err := b.app.switchNS(ns); err != nil {
		b.App().Flash().Err(err)
		return nil
	}
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
	b.GetModel().SetNamespace(ns)
}

func (b *Browser) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, b.app.factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, b.GVR().String())
	if b.Path != "" {
		ctx = context.WithValue(ctx, internal.KeyPath, b.Path)
	}
	if ui.IsLabelSelector(b.CmdBuff().GetText()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(b.CmdBuff().GetText()))
	}
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.CleanseNamespace(b.App().Config.ActiveNamespace()))

	return ctx
}

func (b *Browser) refreshActions() {
	if b.App().Content.Top().Name() != b.Name() {
		return
	}
	aa := ui.KeyActions{
		ui.KeyC:        ui.NewKeyAction("Copy", b.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", b.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", b.refreshCmd, false),
	}

	if b.app.ConOK() {
		b.namespaceActions(aa)
		if !b.app.Config.K9s.IsReadOnly() {
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
	for _, f := range b.bindKeysFn {
		f(aa)
	}
	b.Actions().Add(aa)
	b.app.Menu().HydrateMenu(b.Hints())
}

func (b *Browser) namespaceActions(aa ui.KeyActions) {
	if !b.meta.Namespaced || b.GetTable().Path != "" {
		return
	}
	b.namespaces = make(map[int]string, config.MaxFavoritesNS)
	aa[ui.Key0] = ui.NewKeyAction(client.NamespaceAll, b.switchNamespaceCmd, true)
	b.namespaces[0] = client.NamespaceAll
	index := 1
	for _, ns := range b.app.Config.FavNamespaces() {
		if ns == client.NamespaceAll {
			continue
		}
		aa[ui.NumKeys[index]] = ui.NewKeyAction(ns, b.switchNamespaceCmd, true)
		b.namespaces[index] = ns
		index++
	}
}

func (b *Browser) simpleDelete(selections []string, msg string) {
	dialog.ShowConfirm(b.app.Styles.Dialog(), b.app.Content.Pages, "Confirm Delete", msg, func() {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
		}
		log.Debug().Msgf("SELS %v", selections)
		for _, sel := range selections {
			nuker, ok := b.accessor.(dao.Nuker)
			if !ok {
				b.app.Flash().Errf("Invalid nuker %T", b.accessor)
				continue
			}
			if err := nuker.Delete(context.Background(), sel, nil, false); err != nil {
				b.app.Flash().Errf("Delete failed with `%s", err)
			} else {
				b.app.factory.DeleteForwarder(sel)
			}
			b.GetTable().DeleteMark(sel)
		}
		b.refresh()
	}, func() {})
}

func (b *Browser) resourceDelete(selections []string, msg string) {
	dialog.ShowDelete(b.app.Styles.Dialog(), b.app.Content.Pages, msg, func(propagation *metav1.DeletionPropagation, force bool) {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
		}
		for _, sel := range selections {
			if err := b.GetModel().Delete(b.defaultContext(), sel, propagation, force); err != nil {
				b.app.Flash().Errf("Delete failed with `%s", err)
			} else {
				b.app.factory.DeleteForwarder(sel)
			}
			b.GetTable().DeleteMark(sel)
		}
		b.refresh()
	}, func() {})
}

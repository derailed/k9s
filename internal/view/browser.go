// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
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
	updating   bool
}

// NewBrowser returns a new browser.
func NewBrowser(gvr client.GVR) ResourceViewer {
	return &Browser{
		Table: NewTable(gvr),
	}
}

func (b *Browser) setUpdating(f bool) {
	b.mx.Lock()
	defer b.mx.Unlock()
	b.updating = f
}

func (b *Browser) getUpdating() bool {
	b.mx.RLock()
	defer b.mx.RUnlock()
	return b.updating
}

// Init watches all running pods in given namespace.
func (b *Browser) Init(ctx context.Context) error {
	var err error
	b.meta, err = dao.MetaAccess.MetaFor(b.GVR())
	if err != nil {
		return err
	}
	colorerFn := model1.DefaultColorer
	if r, ok := model.Registry[b.GVR().String()]; ok && r.Renderer != nil {
		colorerFn = r.Renderer.ColorerFunc()
	}
	b.GetTable().SetColorerFn(colorerFn)

	if err = b.Table.Init(ctx); err != nil {
		return err
	}
	ns := client.CleanseNamespace(b.app.Config.ActiveNamespace())
	if dao.IsK8sMeta(b.meta) && b.app.ConOK() {
		if _, e := b.app.factory.CanForResource(ns, b.GVR().String(), client.ListAccess); e != nil {
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

func (b *Browser) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
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
		b.App().Flash().Errf("Watcher failed for %s -- %s", b.GVR(), err)
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

func (b *Browser) SetFilter(s string) {
	b.CmdBuff().SetText(s, "")
}

func (b *Browser) SetLabelFilter(labels map[string]string) {
	b.CmdBuff().SetText(toLabelsStr(labels), "")
	b.GetModel().SetLabelFilter(toLabelsStr(labels))
}

// BufferChanged indicates the buffer was changed.
func (b *Browser) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (b *Browser) BufferCompleted(text, _ string) {
	if internal.IsLabelSelector(text) {
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
	if err := b.GetModel().Refresh(b.GetContext()); err != nil {
		log.Error().Err(err).Msgf("Refresh failed for %s", b.GVR())
	}
	data := b.GetModel().Peek()
	cdata := b.Update(data, b.App().Conn().HasMetrics())
	b.app.QueueUpdateDraw(func() {
		if b.getUpdating() {
			return
		}
		b.setUpdating(true)
		defer b.setUpdating(false)
		b.UpdateUI(cdata, data)
		if b.GetRowCount() > 1 {
			b.App().filterHistory.Push(b.CmdBuff().GetText())
		}

	})
}

func (b *Browser) prepareContext() context.Context {
	ctx := b.defaultContext()

	b.mx.Lock()
	{
		if b.cancelFn != nil {
			b.cancelFn()
		}
		ctx, b.cancelFn = context.WithCancel(ctx)
	}
	b.mx.Unlock()

	if b.contextFn != nil {
		ctx = b.contextFn(ctx)
	}
	if path, ok := ctx.Value(internal.KeyPath).(string); ok && path != "" {
		b.Path = path
	}
	b.mx.Lock()
	{
		b.SetContext(ctx)
	}
	b.mx.Unlock()

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
func (b *Browser) Aliases() map[string]struct{} {
	return aliasesFor(b.meta, b.app.command.AliasesFor(b.meta.Name))
}

// ----------------------------------------------------------------------------
// Model Protocol...

// TableDataChanged notifies view new data is available.
func (b *Browser) TableDataChanged(data *model1.TableData) {
	var cancel context.CancelFunc
	b.mx.RLock()
	cancel = b.cancelFn
	b.mx.RUnlock()

	if !b.app.ConOK() || cancel == nil || !b.app.IsRunning() {
		return
	}

	cdata := b.Update(data, b.app.Conn().HasMetrics())
	b.app.QueueUpdateDraw(func() {
		if b.getUpdating() {
			return
		}
		b.setUpdating(true)
		defer b.setUpdating(false)
		b.refreshActions()
		b.UpdateUI(cdata, data)
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

	v := NewLiveView(b.app, yamlAction, model.NewYAML(b.GVR(), path))
	if err := v.app.inject(v, false); err != nil {
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
		hasFilter := !b.CmdBuff().Empty()
		b.CmdBuff().ClearText(false)
		if hasFilter {
			b.GetModel().SetLabelFilter("")
			b.Refresh()
		}
		return b.App().PrevCmd(evt)
	}

	b.CmdBuff().Reset()
	if internal.IsLabelSelector(b.CmdBuff().GetText()) {
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
	if internal.IsLabelSelector(b.CmdBuff().GetText()) {
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
	f(b.app, b.GetModel(), b.GVR(), path)

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
	describeResource(b.app, b.GetModel(), b.GVR(), path)

	return nil
}

func (b *Browser) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := b.GetSelectedItem()
	if path == "" {
		return evt
	}

	b.Stop()
	defer b.Start()
	if err := editRes(b.app, b.GVR(), path); err != nil {
		b.App().Flash().Err(err)
	}

	return nil
}

func editRes(app *App, gvr client.GVR, path string) error {
	if path == "" {
		return fmt.Errorf("nothing selected %q", path)
	}
	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}
	if gvr.String() == "v1/namespaces" {
		ns = n
	}
	if ok, err := app.Conn().CanI(ns, gvr.String(), n, client.PatchAccess); !ok || err != nil {
		return fmt.Errorf("current user can't edit resource %s", gvr)
	}

	args := make([]string, 0, 10)
	args = append(args, "edit")
	args = append(args, gvr.FQN(n))
	if ns != client.BlankNamespace {
		args = append(args, "-n", ns)
	}
	if err := runK(app, shellOpts{clear: true, args: args}); err != nil {
		app.Flash().Errf("Edit command failed: %s", err)
	}

	return nil
}

func (b *Browser) switchNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	i, err := strconv.Atoi(string(evt.Rune()))
	if err != nil {
		log.Error().Err(err).Msgf("Fail to switch namespace")
		return nil
	}
	ns := b.namespaces[i]

	auth, err := b.App().factory.Client().CanI(ns, b.GVR().String(), "", client.ListAccess)
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
	b.SelectRow(1, 0, true)
	b.app.CmdBuff().Reset()
	if err := b.app.Config.SetActiveNamespace(b.GetModel().GetNamespace()); err != nil {
		log.Error().Err(err).Msg("Config save NS failed!")
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
	ctx = context.WithValue(ctx, internal.KeyGVR, b.GVR())
	ctx = context.WithValue(ctx, internal.KeyPath, b.Path)
	if internal.IsLabelSelector(b.CmdBuff().GetText()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(b.CmdBuff().GetText()))
	}
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.CleanseNamespace(b.App().Config.ActiveNamespace()))
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, b.app.factory.Client().HasMetrics())

	return ctx
}

func (b *Browser) refreshActions() {
	if b.App().Content.Top() != nil && b.App().Content.Top().Name() != b.Name() {
		return
	}
	aa := ui.NewKeyActionsFromMap(ui.KeyMap{
		ui.KeyC:        ui.NewKeyAction("Copy", b.cpCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", b.enterCmd, false),
		tcell.KeyCtrlR: ui.NewKeyAction("Refresh", b.refreshCmd, false),
	})

	if b.app.ConOK() {
		b.namespaceActions(aa)
		if !b.app.Config.K9s.IsReadOnly() {
			if client.Can(b.meta.Verbs, "edit") {
				aa.Add(ui.KeyE, ui.NewKeyActionWithOpts("Edit", b.editCmd,
					ui.ActionOpts{
						Visible:   true,
						Dangerous: true,
					}))
			}
			if client.Can(b.meta.Verbs, "delete") {
				aa.Add(tcell.KeyCtrlD, ui.NewKeyActionWithOpts("Delete", b.deleteCmd,
					ui.ActionOpts{
						Visible:   true,
						Dangerous: true,
					}))
			}
		} else {
			b.Actions().ClearDanger()
		}
	}
	if !dao.IsK9sMeta(b.meta) {
		aa.Add(ui.KeyY, ui.NewKeyAction(yamlAction, b.viewCmd, true))
		aa.Add(ui.KeyD, ui.NewKeyAction("Describe", b.describeCmd, true))
	}
	for _, f := range b.bindKeysFn {
		f(aa)
	}
	b.Actions().Merge(aa)

	if err := pluginActions(b, b.Actions()); err != nil {
		log.Warn().Msgf("Plugins load failed: %s", err)
		b.app.Logo().Warn("Plugins load failed!")
	}
	if err := hotKeyActions(b, b.Actions()); err != nil {
		log.Warn().Msgf("Hotkeys load failed: %s", err)
		b.app.Logo().Warn("HotKeys load failed!")
	}
	b.app.Menu().HydrateMenu(b.Hints())
}

func (b *Browser) namespaceActions(aa *ui.KeyActions) {
	if !b.meta.Namespaced || b.GetTable().Path != "" {
		return
	}
	aa.Add(ui.KeyN, ui.NewKeyAction("Copy Namespace", b.cpNsCmd, false))

	b.namespaces = make(map[int]string, data.MaxFavoritesNS)
	aa.Add(ui.Key0, ui.NewKeyAction(client.NamespaceAll, b.switchNamespaceCmd, true))
	b.namespaces[0] = client.NamespaceAll
	index := 1
	favNamespaces := b.app.Config.FavNamespaces()
	for _, ns := range favNamespaces {
		if ns == client.NamespaceAll {
			continue
		}
		if numKey, ok := ui.NumKeys[index]; ok {
			aa.Add(numKey, ui.NewKeyAction(ns, b.switchNamespaceCmd, true))
			b.namespaces[index] = ns
			index++
		} else {
			log.Warn().Msgf("No number key available for favorite namespace %s (%d of %d). Skipping...", ns, index, len(favNamespaces))
			break
		}
	}
}

func (b *Browser) simpleDelete(selections []string, msg string) {
	dialog.ShowConfirm(b.app.Styles.Dialog(), b.app.Content.Pages, "Confirm Delete", msg, func() {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR().R())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
		}
		for _, sel := range selections {
			nuker, ok := b.accessor.(dao.Nuker)
			if !ok {
				b.app.Flash().Errf("Invalid nuker %T", b.accessor)
				continue
			}
			if err := nuker.Delete(context.Background(), sel, nil, dao.DefaultGrace); err != nil {
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
	okFn := func(propagation *metav1.DeletionPropagation, force bool) {
		b.ShowDeleted()
		if len(selections) > 1 {
			b.app.Flash().Infof("Delete %d marked %s", len(selections), b.GVR())
		} else {
			b.app.Flash().Infof("Delete resource %s %s", b.GVR(), selections[0])
		}
		for _, sel := range selections {
			grace := dao.DefaultGrace
			if force {
				grace = dao.ForceGrace
			}
			if err := b.GetModel().Delete(b.defaultContext(), sel, propagation, grace); err != nil {
				b.app.Flash().Errf("Delete failed with `%s", err)
			} else {
				b.app.factory.DeleteForwarder(sel)
			}
			b.GetTable().DeleteMark(sel)
		}
		b.refresh()
	}
	dialog.ShowDelete(b.app.Styles.Dialog(), b.app.Content.Pages, msg, okFn, func() {})
}

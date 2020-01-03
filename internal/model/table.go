package model

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
)

// TableListener represents a table model listener.
type TableListener interface {
	// TableDataChanged notifies the model data changed.
	TableDataChanged(render.TableData)

	// TableLoadFailed notifies the load failed.
	TableLoadFailed(error)
}

// Table represents a table model.
type Table struct {
	gvr         string
	namespace   string
	data        *render.TableData
	listeners   []TableListener
	inUpdate    int32
	refreshRate time.Duration
}

// NewTable returns a new table model.
func NewTable(gvr string) *Table {
	return &Table{
		gvr:         gvr,
		data:        render.NewTableData(),
		refreshRate: 2 * time.Second,
	}
}

// Watch initiates model updates.
func (t *Table) Watch(ctx context.Context) {
	t.Refresh(ctx)
	go t.updater(ctx)
}

// Get returns a resource instance if found, else an error.
func (t *Table) Get(ctx context.Context, path string) (runtime.Object, error) {
	meta := t.resourceMeta()
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.Model.Init(t.namespace, t.gvr, factory)

	return meta.Model.Get(ctx, path)
}

// Refresh update the model now.
func (t *Table) Refresh(ctx context.Context) {
	t.refresh(ctx)
}

// GetNamespace returns the model namespace.
func (t *Table) GetNamespace() string {
	return t.namespace
}

// SetNamespace sets up model namespace.
func (t *Table) SetNamespace(ns string) {
	t.namespace = ns
	t.data.Clear()
}

// SetRefreshRate sets model refresh duration.
func (t *Table) SetRefreshRate(d time.Duration) {
	t.refreshRate = d
}

// ClusterWide checks if resource is scope for all namespaces.
func (t *Table) ClusterWide() bool {
	return t.namespace == render.AllNamespaces
}

// InNamespace checks if current namespace matches desired namespace.
func (t *Table) InNamespace(ns string) bool {
	return t.namespace == ns
}

// Empty return true if no model data.
func (t *Table) Empty() bool {
	return len(t.data.RowEvents) == 0
}

// Peek returns model data.
func (t *Table) Peek() render.TableData {
	return *t.data
}

func (t *Table) updater(ctx context.Context) {
	defer log.Debug().Msgf("Model canceled -- %q", t.gvr)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.refreshRate):
			t.refresh(ctx)
		}
	}
}

func (t *Table) refresh(ctx context.Context) {
	log.Debug().Msgf("RECONCILING")
	if !atomic.CompareAndSwapInt32(&t.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return
	}
	defer atomic.StoreInt32(&t.inUpdate, 0)

	if err := t.reconcile(ctx); err != nil {
		log.Error().Err(err).Msg("Reconcile failed")
		t.fireTableLoadFailed(err)
		return
	}
	t.fireTableChanged(*t.data)
}

// AddListener adds a new model listener.
func (t *Table) AddListener(l TableListener) {
	t.listeners = append(t.listeners, l)
}

// RemoveListener delete a listener from the list.
func (t *Table) RemoveListener(l TableListener) {
	victim := -1
	for i, lis := range t.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		t.listeners = append(t.listeners[:victim], t.listeners[victim+1:]...)
	}
}

func (t *Table) fireTableChanged(data render.TableData) {
	for _, l := range t.listeners {
		l.TableDataChanged(data)
	}
}

func (t *Table) fireTableLoadFailed(err error) {
	for _, l := range t.listeners {
		l.TableLoadFailed(err)
	}
}

func (t *Table) list(ctx context.Context, l Lister) ([]runtime.Object, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	l.Init(t.namespace, t.gvr, factory)

	return l.List(ctx)
}

func (t *Table) resourceMeta() ResourceMeta {
	meta, ok := Registry[t.gvr]
	if !ok {
		log.Debug().Msgf("Resource %s not found in registry. Going generic!", t.gvr)
		meta = ResourceMeta{
			Model:    &Generic{},
			Renderer: &render.Generic{},
		}
	}
	if meta.Model == nil {
		meta.Model = &Resource{}
	}

	return meta
}

func (t *Table) reconcile(ctx context.Context) error {
	defer func(t time.Time) {
		log.Debug().Msgf("RECONCILE elapsed %v", time.Since(t))
	}(time.Now())

	meta := t.resourceMeta()
	oo, err := t.list(ctx, meta.Model)
	if err != nil {
		return err
	}
	log.Debug().Msgf("LIST returned %d rows", len(oo))

	rows := make(render.Rows, len(oo))
	if err := meta.Model.Hydrate(oo, rows, meta.Renderer); err != nil {
		return err
	}

	t.data.Mutex.Lock()
	defer t.data.Mutex.Unlock()
	// if labelSelector in place might as well clear the model data.
	sel, ok := ctx.Value(internal.KeyLabels).(string)
	if ok && sel != "" {
		t.data.Clear()
	}
	t.data.Update(rows)
	t.data.Namespace, t.data.Header = t.namespace, meta.Renderer.Header(t.namespace)
	log.Debug().Msgf("TABLE_DATA returns %d rows", len(t.data.RowEvents))

	return nil
}

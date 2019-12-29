package model

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
)

const refreshRate = 1 * time.Second

type TableListener interface {
	TableDataChanged(render.TableData)
	TableLoadFailed(error)
}

type Table struct {
	gvr         string
	namespace   string
	data        render.TableData
	listeners   []TableListener
	inUpdate    int32
	refreshRate time.Duration
}

// NewTable returns a new table model.
func NewTable(gvr string) *Table {
	return &Table{
		gvr:         gvr,
		data:        render.TableData{},
		refreshRate: 2 * time.Second,
	}
}

// Watch initiates model updates.
func (t *Table) Watch(ctx context.Context) {
	t.Refresh(ctx)
	go t.updater(ctx)
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
	return t.data
}

func (t *Table) updater(ctx context.Context) {
	defer log.Debug().Msgf("Model canceled -- %q", t.gvr)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshRate):
			t.refresh(ctx)
		}
	}
}

func (t *Table) refresh(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&t.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return
	}
	defer atomic.StoreInt32(&t.inUpdate, 0)

	if err := t.reconcile(ctx); err != nil {
		log.Error().Err(err).Msg("Reconcile failed")
		t.fireTableLoadFailed(err)
	}
	t.fireTableChanged(t.data)
}

// AddListener adds a new model listener.
func (t *Table) AddListener(l TableListener) {
	t.listeners = append(t.listeners, l)
	t.fireTableChanged(t.data)
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

func (t *Table) reconcile(ctx context.Context) error {
	log.Debug().Msgf("GOROUTINE %d", runtime.NumGoroutine())

	factory, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok {
		return fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	m, ok := Registry[string(t.gvr)]
	if !ok {
		log.Warn().Msgf("Resource %s not found in registry. Going generic!", t.gvr)
		m = ResourceMeta{
			Model:    &Generic{},
			Renderer: &render.Generic{},
		}
	}

	if m.Model == nil {
		m.Model = &Resource{}
	}
	m.Model.Init(t.namespace, string(t.gvr), factory)
	oo, err := m.Model.List(ctx)
	if err != nil {
		return err
	}

	rows := make(render.Rows, len(oo))
	if err := m.Model.Hydrate(oo, rows, m.Renderer); err != nil {
		return err
	}
	t.data.Update(rows)
	t.data.Namespace, t.data.Header = t.namespace, m.Renderer.Header(t.namespace)

	return nil
}

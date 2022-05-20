package model

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const initRefreshRate = 300 * time.Millisecond

// TableListener represents a table model listener.
type TableListener interface {
	// TableDataChanged notifies the model data changed.
	TableDataChanged(render.TableData)

	// TableLoadFailed notifies the load failed.
	TableLoadFailed(error)
}

// Table represents a table model.
type Table struct {
	gvr         client.GVR
	namespace   string
	data        *render.TableData
	listeners   []TableListener
	inUpdate    int32
	refreshRate time.Duration
	instance    string
	mx          sync.RWMutex
	labelFilter string
}

// NewTable returns a new table model.
func NewTable(gvr client.GVR) *Table {
	return &Table{
		gvr:         gvr,
		data:        render.NewTableData(),
		refreshRate: 2 * time.Second,
	}
}

// SetLabelFilter sets the labels filter.
func (t *Table) SetLabelFilter(f string) {
	t.mx.Lock()
	t.labelFilter = f
	t.mx.Unlock()
}

// SetInstance sets a single entry table.
func (t *Table) SetInstance(path string) {
	t.instance = path
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
		t.mx.Lock()
		defer t.mx.Unlock()
		t.listeners = append(t.listeners[:victim], t.listeners[victim+1:]...)
	}
}

// Watch initiates model updates.
func (t *Table) Watch(ctx context.Context) error {
	if err := t.refresh(ctx); err != nil {
		return err
	}
	go t.updater(ctx)

	return nil
}

// Refresh updates the table content.
func (t *Table) Refresh(ctx context.Context) error {
	return t.refresh(ctx)
}

// Get returns a resource instance if found, else an error.
func (t *Table) Get(ctx context.Context, path string) (runtime.Object, error) {
	meta, err := getMeta(ctx, t.gvr)
	if err != nil {
		return nil, err
	}

	return meta.DAO.Get(ctx, path)
}

// Delete deletes a resource.
func (t *Table) Delete(ctx context.Context, path string, propagation *metav1.DeletionPropagation, force bool) error {
	meta, err := getMeta(ctx, t.gvr)
	if err != nil {
		return err
	}

	nuker, ok := meta.DAO.(dao.Nuker)
	if !ok {
		return fmt.Errorf("no nuker for %q", meta.DAO.GVR())
	}

	return nuker.Delete(path, propagation, force)
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

// InNamespace checks if current namespace matches desired namespace.
func (t *Table) InNamespace(ns string) bool {
	return len(t.data.RowEvents) > 0 && t.namespace == ns
}

// SetRefreshRate sets model refresh duration.
func (t *Table) SetRefreshRate(d time.Duration) {
	t.refreshRate = d
}

// ClusterWide checks if resource is scope for all namespaces.
func (t *Table) ClusterWide() bool {
	return client.IsClusterWide(t.namespace)
}

// Empty returns true if no model data.
func (t *Table) Empty() bool {
	return len(t.data.RowEvents) == 0
}

// Count returns the row count.
func (t *Table) Count() int {
	return len(t.data.RowEvents)
}

// Peek returns model data.
func (t *Table) Peek() render.TableData {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.data.Clone()
}

func (t *Table) updater(ctx context.Context) {
	defer log.Debug().Msgf("TABLE-UPDATER canceled -- %q", t.gvr)

	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval, bf.MaxElapsedTime = initRefreshRate, maxReaderRetryInterval
	rate := initRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(rate):
			rate = t.refreshRate
			err := backoff.Retry(func() error {
				return t.refresh(ctx)
			}, backoff.WithContext(bf, ctx))
			if err != nil {
				log.Error().Err(err).Msgf("Retry failed")
				t.fireTableLoadFailed(err)
				return
			}
		}
	}
}

func (t *Table) refresh(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&t.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return nil
	}
	defer atomic.StoreInt32(&t.inUpdate, 0)

	if err := t.reconcile(ctx); err != nil {
		return err
	}
	t.fireTableChanged(t.Peek())

	return nil
}

func (t *Table) list(ctx context.Context, a dao.Accessor) ([]runtime.Object, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	a.Init(factory, t.gvr)

	ns := client.CleanseNamespace(t.namespace)
	if client.IsClusterScoped(t.namespace) {
		ns = client.AllNamespaces
	}

	return a.List(ctx, ns)
}

func (t *Table) reconcile(ctx context.Context) error {
	t.mx.Lock()
	defer t.mx.Unlock()
	meta := resourceMeta(t.gvr)
	if t.labelFilter != "" {
		ctx = context.WithValue(ctx, internal.KeyLabels, t.labelFilter)
	}
	var (
		oo  []runtime.Object
		err error
	)
	if t.instance == "" {
		oo, err = t.list(ctx, meta.DAO)
	} else {
		o, e := t.Get(ctx, t.instance)
		oo, err = []runtime.Object{o}, e
	}
	if err != nil {
		return err
	}

	var rows render.Rows
	if len(oo) > 0 {
		if meta.Renderer.IsGeneric() {
			table, ok := oo[0].(*metav1beta1.Table)
			if !ok {
				return fmt.Errorf("expecting a meta table but got %T", oo[0])
			}
			rows = make(render.Rows, len(table.Rows))
			if err := genericHydrate(t.namespace, table, rows, meta.Renderer); err != nil {
				return err
			}
		} else {
			rows = make(render.Rows, len(oo))
			if err := hydrate(t.namespace, oo, rows, meta.Renderer); err != nil {
				return err
			}
		}
	}

	// if labelSelector in place might as well clear the model data.
	sel, ok := ctx.Value(internal.KeyLabels).(string)
	if ok && sel != "" {
		t.data.Clear()
	}
	t.data.Update(rows)
	t.data.SetHeader(t.namespace, meta.Renderer.Header(t.namespace))

	if len(t.data.Header) == 0 {
		return fmt.Errorf("fail to list resource %s", t.gvr)
	}

	return nil
}

func (t *Table) fireTableChanged(data render.TableData) {
	t.mx.RLock()
	defer t.mx.RUnlock()

	for _, l := range t.listeners {
		l.TableDataChanged(data)
	}
}

func (t *Table) fireTableLoadFailed(err error) {
	for _, l := range t.listeners {
		l.TableLoadFailed(err)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func hydrate(ns string, oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, ns, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

type Generic interface {
	SetTable(*metav1beta1.Table)
	Header(string) render.Header
	Render(interface{}, string, *render.Row) error
}

func genericHydrate(ns string, table *metav1beta1.Table, rr render.Rows, re Renderer) error {
	gr, ok := re.(Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer but got %T", re)
	}
	gr.SetTable(table)
	for i, row := range table.Rows {
		if err := gr.Render(row, ns, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

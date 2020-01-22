package model

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
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
	gvr         string
	namespace   string
	data        *render.TableData
	listeners   []TableListener
	inUpdate    int32
	refreshRate time.Duration
	instance    string
}

// NewTable returns a new table model.
func NewTable(gvr string) *Table {
	return &Table{
		gvr:         gvr,
		data:        render.NewTableData(),
		refreshRate: 2 * time.Second,
	}
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
		t.listeners = append(t.listeners[:victim], t.listeners[victim+1:]...)
	}
}

// Watch initiates model updates.
func (t *Table) Watch(ctx context.Context) {
	t.Refresh(ctx)
	go t.updater(ctx)
}

// Get returns a resource instance if found, else an error.
func (t *Table) Get(ctx context.Context, path string) (runtime.Object, error) {
	meta, err := t.getMeta(ctx)
	if err != nil {
		return nil, err
	}

	return meta.DAO.Get(ctx, path)
}

// Delete deletes a resource.
func (t *Table) Delete(ctx context.Context, path string, cascade, force bool) error {
	meta, err := t.getMeta(ctx)
	if err != nil {
		return err
	}

	nuker, ok := meta.DAO.(dao.Nuker)
	if !ok {
		return fmt.Errorf("no nuker for %q", meta.DAO.GVR())
	}

	return nuker.Delete(path, cascade, force)
}

// Describe describes a given resource.
func (t *Table) Describe(ctx context.Context, path string) (string, error) {
	meta, err := t.getMeta(ctx)
	if err != nil {
		return "", err
	}

	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	return desc.Describe(path)
}

// ToYAML returns a resource yaml.
func (t *Table) ToYAML(ctx context.Context, path string) (string, error) {
	meta, err := t.getMeta(ctx)
	if err != nil {
		return "", err
	}

	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	return desc.ToYAML(path)
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

	rate := initRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(rate):
			rate = t.refreshRate
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
		return
	}
	t.fireTableChanged(*t.data)
}

func (t *Table) list(ctx context.Context, a dao.Accessor) ([]runtime.Object, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	a.Init(factory, client.NewGVR(t.gvr))

	ns := client.CleanseNamespace(t.namespace)
	if client.IsClusterScoped(t.namespace) {
		ns = client.AllNamespaces
	}
	return a.List(ctx, ns)
}

func (t *Table) reconcile(ctx context.Context) error {
	meta := t.resourceMeta()
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
	if _, ok := meta.Renderer.(*render.Generic); ok {
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

	t.data.Mutex.Lock()
	defer t.data.Mutex.Unlock()
	// if labelSelector in place might as well clear the model data.
	sel, ok := ctx.Value(internal.KeyLabels).(string)
	if ok && sel != "" {
		t.data.Clear()
	}
	t.data.Update(rows)
	t.data.Namespace, t.data.Header = t.namespace, meta.Renderer.Header(t.namespace)

	return nil
}

func (t *Table) getMeta(ctx context.Context) (ResourceMeta, error) {
	meta := t.resourceMeta()
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return ResourceMeta{}, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.DAO.Init(factory, client.NewGVR(t.gvr))

	return meta, nil
}

func (t *Table) resourceMeta() ResourceMeta {
	meta, ok := Registry[t.gvr]
	if !ok {
		log.Debug().Msgf("Resource %s not found in registry. Going generic!", t.gvr)
		meta = ResourceMeta{
			DAO:      &dao.Table{},
			Renderer: &render.Generic{},
		}
	}
	if meta.DAO == nil {
		meta.DAO = &dao.Resource{}
	}

	return meta
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

func genericHydrate(ns string, table *metav1beta1.Table, rr render.Rows, re Renderer) error {
	gr, ok := re.(*render.Generic)
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

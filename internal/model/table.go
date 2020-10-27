package model

import (
	"context"
	"fmt"
	"github.com/derailed/k9s/internal/config"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/util/fieldpath"
	"sync"
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
	gvr          client.GVR
	namespace    string
	data         *render.TableData
	listeners    []TableListener
	inUpdate     int32
	refreshRate  time.Duration
	instance     string
	mx           sync.RWMutex
	labelFilter  string
	modelSetting *config.ModelSetting
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
	t.labelFilter = f
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
func (t *Table) Watch(ctx context.Context) {
	t.refresh(ctx)
	go t.updater(ctx)
}

// Refresh updates the table content.
func (t *Table) Refresh(ctx context.Context) {
	t.refresh(ctx)
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
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.data.Clone()
}

func (t *Table) updater(ctx context.Context) {
	defer log.Debug().Msgf("TABLE-MODEL canceled -- %q", t.gvr)

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
		log.Error().Err(err).Msgf("reconcile failed %q::%q", t.gvr, t.instance)
		t.fireTableLoadFailed(err)
		return
	}
	t.fireTableChanged(t.Peek())
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
	meta := t.resourceMeta()
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

	currentModelSetting := t.modelSetting
	header := customizeHeader(meta.Renderer.Header(t.namespace), currentModelSetting)

	var rows render.Rows
	if len(oo) > 0 {
		if _, ok := meta.Renderer.(*render.Generic); ok {
			table, ok := oo[0].(*metav1beta1.Table)
			if !ok {
				return fmt.Errorf("expecting a meta table but got %T", oo[0])
			}
			rows = make(render.Rows, len(table.Rows))
			if err := genericHydrate(t.namespace, table, rows, meta.Renderer, currentModelSetting); err != nil {
				return err
			}
		} else {
			rows = make(render.Rows, len(oo))
			if err := hydrate(t.namespace, oo, rows, meta.Renderer, currentModelSetting); err != nil {
				return err
			}
		}
	}

	t.mx.Lock()
	defer t.mx.Unlock()
	// if labelSelector in place might as well clear the model data.
	sel, ok := ctx.Value(internal.KeyLabels).(string)
	if ok && sel != "" {
		t.data.Clear()
	}
	t.data.Update(rows)
	t.data.SetHeader(t.namespace, header)

	if len(t.data.Header) == 0 {
		return fmt.Errorf("fail to list resource %s", t.gvr)
	}

	return nil
}

func customizeHeader(header render.Header, setting *config.ModelSetting) render.Header {
	if setting == nil {
		return header
	}

	for _, col := range setting.Columns {
		header = append(header, render.HeaderColumn{
			Name: col.Name,
			// BOZO fill the rest of the settings
		})
	}
	return header
}

func (t *Table) getMeta(ctx context.Context) (ResourceMeta, error) {
	meta := t.resourceMeta()
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return ResourceMeta{}, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.DAO.Init(factory, t.gvr)

	return meta, nil
}

func (t *Table) resourceMeta() ResourceMeta {
	meta, ok := Registry[t.gvr.String()]
	if !ok {
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

func (t *Table) Init(ctx context.Context) {
	if cfg, ok := ctx.Value(internal.KeyModelConfig).(*config.CustomModel); ok && cfg != nil {
		cfg.AddListener(t.gvr.String(), t)
	}
}

func (t *Table) ModelSettingsChanged(newSetting config.ModelSetting) {
	t.modelSetting = &newSetting
	// BOZO should tell model to reload, but missing proper context for that
}

// ----------------------------------------------------------------------------
// Helpers...

func hydrate(ns string, oo []runtime.Object, rr render.Rows, re Renderer, setting *config.ModelSetting) error {
	for i, o := range oo {
		if err := re.Render(o, ns, &rr[i]); err != nil {
			return err
		}
		customizeRow(setting, o, &rr[i])
	}

	return nil
}

func customizeRow(setting *config.ModelSetting, o interface{}, row *render.Row) {
	if setting == nil {
		return
	}

	if u, ok := tryExtractObject(o); ok {
		for _, customColumn := range setting.Columns {
			row.Fields = append(row.Fields, extractFieldOrErrorMsg(u, customColumn.FieldPath))
		}
	} else {
		for range setting.Columns {
			row.Fields = append(row.Fields, "<unsupported type>")
		}
	}
}

func tryExtractObject(o interface{}) (v1.Object, bool) {
	switch oo := o.(type) {
	case v1.Object:
		return oo, true
	case render.RenderableRaw:
		return oo.Object(), true
	case v1.TableRow:
		// TODO find a way to test this
		if ooo, ok := oo.Object.Object.(v1.Object); ok {
			return ooo, true
		} else {
			return nil, false
		}
	default:
		return nil, false
	}
}

func extractFieldOrErrorMsg(obj v1.Object, fieldPath string) string {
	fieldValue, err := fieldpath.ExtractFieldPathAsString(obj, fieldPath)
	if err != nil {
		return fmt.Sprintf("err: %s", err.Error())
	}
	return fieldValue
}

func genericHydrate(ns string, table *metav1beta1.Table, rr render.Rows, re Renderer, setting *config.ModelSetting) error {
	gr, ok := re.(*render.Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer but got %T", re)
	}
	gr.SetTable(table)
	for i, row := range table.Rows {
		if err := gr.Render(row, ns, &rr[i]); err != nil {
			return err
		}
		customizeRow(setting, row, &rr[i])
	}

	return nil
}

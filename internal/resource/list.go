package resource

import (
	"context"
	"reflect"

	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/watch"
)

type list struct {
	namespace, name string
	verbs           int
	resource        Resource
	cache           render.RowEvents
	fieldSelector   string
	labelSelector   string
	header          render.HeaderRow
}

// NewList returns a new resource list.
func NewList(ns, name string, res Resource, verbs int) *list {
	return &list{
		namespace: ns,
		name:      name,
		verbs:     verbs,
		resource:  res,
	}
}

func (l *list) HasSelectors() bool {
	return l.fieldSelector != "" || l.labelSelector != ""
}

// SetFieldSelector narrows down resource query given fields selection.
func (l *list) SetFieldSelector(s string) {
	l.fieldSelector = s
}

// SetLabelSelector narrows down resource query via labels selections.
func (l *list) SetLabelSelector(s string) {
	l.labelSelector = s
}

// Access check access control on a given resource.
func (l *list) Access(f int) bool {
	return l.verbs&f == f
}

// Access check access control on a given resource.
func (l *list) GetAccess() int {
	return l.verbs
}

// Access check access control on a given resource.
func (l *list) SetAccess(f int) {
	l.verbs = f
}

// Namespaced checks if k8s resource is namespaced.
func (l *list) Namespaced() bool {
	return l.namespace != NotNamespaced
}

// IsClusterWide returns true if the resource is cluster scoped.
func (l *list) IsCluterWide() bool {
	return l.namespace == render.ClusterWide
}

// AllNamespaces checks if this resource spans all namespaces.
func (l *list) AllNamespaces() bool {
	return l.namespace == AllNamespaces
}

// GetNamespace associated with the resource.
func (l *list) GetNamespace() string {
	if !l.Access(NamespaceAccess) {
		l.namespace = NotNamespaced
	}

	return l.namespace
}

// SetNamespace updates the namespace on the list. Default ns is "" for all
// namespaces.
func (l *list) SetNamespace(n string) {
	if !l.Namespaced() {
		return
	}

	if n == AllNamespace {
		n = AllNamespaces
	}
	if l.namespace == n {
		return
	}
	l.cache = nil
	if l.Access(NamespaceAccess) {
		l.namespace = n
		if n == AllNamespace {
			l.namespace = AllNamespaces
		}
	}
}

// GetName returns the kubernetes resource name.
func (l *list) GetName() string {
	return l.name
}

// Resource returns a resource api connection.
func (l *list) Resource() Resource {
	return l.resource
}

// Cache tracks previous resource state.
func (l *list) Data() render.TableData {
	return render.TableData{
		Header:    l.header,
		RowEvents: l.cache,
		Namespace: l.namespace,
	}
}

// BOZO!!
// func (l *list) load(informer *wa.Informer, ns string) (Columnars, error) {
// 	rr, err := informer.List(l.name, ns, metav1.ListOptions{
// 		FieldSelector: l.fieldSelector,
// 		LabelSelector: l.labelSelector,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	items := make(Columnars, 0, len(rr))
// 	for _, r := range rr {
// 		res, err := l.fetchResource(informer, r, ns)
// 		if err != nil {
// 			return nil, err
// 		}
// 		items = append(items, res)
// 	}

// 	return items, nil
// }

// BOZO!!
// func (l *list) fetchResource(informer *wa.Informer, r interface{}, ns string) (Columnar, error) {
// 	res, err := l.resource.New(r)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch o := r.(type) {
// 	case *v1.Node:
// 		fqn := MetaFQN(o.ObjectMeta)
// 		nmx, err := informer.Get(wa.NodeMXIndex, fqn, metav1.GetOptions{})
// 		if err != nil {
// 			return res, err
// 		}
// 		res.SetNodeMetrics(nmx.(*mv1beta1.NodeMetrics))
// 	case *v1.Pod:
// 		fqn := MetaFQN(o.ObjectMeta)
// 		pmx, err := informer.Get(wa.PodMXIndex, fqn, metav1.GetOptions{})
// 		if err != nil {
// 			return res, err
// 		}
// 		res.SetPodMetrics(pmx.(*mv1beta1.PodMetrics))
// 	case v1.Container:
// 		pmx, err := informer.Get(wa.PodMXIndex, ns, metav1.GetOptions{})
// 		if err != nil {
// 			return res, err
// 		}
// 		res.SetPodMetrics(pmx.(*mv1beta1.PodMetrics))
// 	default:
// 		return res, fmt.Errorf("No informer matched %s:%s", l.name, ns)
// 	}

// 	return res, nil
// }

// Reconcile previous vs current state and emits delta events.
func (l *list) Reconcile(ctx context.Context, gvr string) error {
	panic("NYI")
	// path := ctx.Value(internal.KeySelection).(string)

	// log.Debug().Msgf("Reconcile %q in path %q", gvr, path)
	// ns := l.namespace
	// if path != "" {
	// 	ns = path
	// }

	// factory, ok := ctx.Value(internal.KeyFactory).(*w.Factory)
	// if !ok {
	// 	return errors.New("no factory found in context")
	// }
	// m, ok := model.Registry[gvr]
	// if !ok {
	// 	log.Warn().Msgf("Resource %s not found in registry. Going generic!", gvr)
	// 	m = model.ResourceMeta{
	// 		Model:    &model.Generic{},
	// 		Renderer: &render.Generic{},
	// 	}
	// }
	// if m.Model == nil {
	// 	m.Model = &model.Resource{}
	// }
	// m.Model.Init(ns, gvr, factory)

	// if l.labelSelector != "" {
	// 	ctx = context.WithValue(ctx, internal.KeyLabels, l.labelSelector)
	// }
	// if l.fieldSelector != "" {
	// 	ctx = context.WithValue(ctx, internal.KeyFields, l.fieldSelector)
	// }
	// oo, err := m.Model.List(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Debug().Msgf("Model returned [%d] items", len(oo))
	// rows := make(render.Rows, len(oo))
	// if err := m.Model.Hydrate(oo, rows, m.Renderer); err != nil {
	// 	panic(err)
	// }
	// l.update(ns, rows)
	// l.header = m.Renderer.Header(ns)

	// return nil
}

func (l *list) update(ns string, rows render.Rows) {
	cacheEmpty := len(l.cache) == 0
	kk := make([]string, 0, len(rows))
	for _, row := range rows {
		kk = append(kk, row.ID)
		if cacheEmpty {
			l.cache = append(l.cache, render.NewRowEvent(render.EventAdd, row))
			continue
		}
		if index, ok := l.cache.FindIndex(row.ID); ok {
			delta := render.NewDeltaRow(l.cache[index].Row, row, true)
			if delta.IsBlank() {
				l.cache[index].Kind, l.cache[index].Deltas = render.EventUnchanged, delta
			} else {
				l.cache[index] = render.NewDeltaRowEvent(row, delta)
			}
			continue
		}
		l.cache = append(l.cache, render.NewRowEvent(render.EventAdd, row))
	}

	if cacheEmpty {
		return
	}
	l.ensureDeletes(kk)
}

// BOZO!!
// // Reconcile previous vs current state and emits delta events.
// func (l *list) Reconcile(informer *wa.Informer, path *string) error {
// 	ns := l.namespace
// 	if path != nil {
// 		ns = *path
// 	}
// 	log.Debug().Msgf("Reconcile in NS %q -- %#v", ns, path)
// 	if items, err := l.load(informer, ns); err == nil {
// 		l.update(items)
// 		return nil
// 	}

// 	opts := metav1.ListOptions{
// 		LabelSelector: l.labelSelector,
// 		FieldSelector: l.fieldSelector,
// 	}
// 	items, err := l.resource.List(l.namespace, opts)
// 	if err != nil {
// 		return err
// 	}
// 	l.update(items)

// 	return nil
// }

// func (l *list) update(items Columnars) {
// 	first := len(l.cache) == 0
// 	kk := make([]string, 0, len(items))
// 	for _, i := range items {
// 		kk = append(kk, i.Name())
// 		ff := i.Fields(l.namespace)
// 		if first {
// 			l.cache[i.Name()] = newRowEvent(New, ff, make(Row, len(ff)))
// 			continue
// 		}
// 		dd := make(Row, len(ff))
// 		a := watch.Added
// 		if evt, ok := l.cache[i.Name()]; ok {
// 			a = computeDeltas(evt, ff[:len(ff)-1], dd)
// 		}
// 		l.cache[i.Name()] = newRowEvent(a, ff, dd)
// 	}

// 	if first {
// 		return
// 	}
// 	l.ensureDeletes(kk)
// }

// EnsureDeletes delete items in cache that are no longer valid.
func (l *list) ensureDeletes(newKeys []string) {
	for _, re := range l.cache {
		var found bool
		for i, key := range newKeys {
			if key == re.Row.ID {
				found = true
				newKeys = append(newKeys[:i], newKeys[i+1:]...)
				break
			}
		}
		if !found {
			l.cache = l.cache.Delete(re.Row.ID)
		}
	}
}

// Helpers...

func computeDeltas(evt *RowEvent, newRow, deltas Row) watch.EventType {
	oldRow := evt.Fields[:len(evt.Fields)-1]
	a := Unchanged
	if !reflect.DeepEqual(oldRow, newRow) {
		for i, field := range oldRow {
			if field != newRow[i] {
				deltas[i] = field
			}
		}
		a = watch.Modified
	}
	return a
}

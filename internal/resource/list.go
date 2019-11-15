package resource

import (
	"fmt"
	"reflect"

	wa "github.com/derailed/k9s/internal/watch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// GetAccess set if resource can be fetched.
	GetAccess = 1 << iota
	// ListAccess set if resource can be listed.
	ListAccess
	// EditAccess set if resource can be edited.
	EditAccess
	// DeleteAccess set if resource can be deleted.
	DeleteAccess
	// ViewAccess set if resource can be viewed.
	ViewAccess
	// NamespaceAccess set if namespaced resource.
	NamespaceAccess
	// DescribeAccess set if resource can be described.
	DescribeAccess
	// SwitchAccess set if resource can be switched (Context).
	SwitchAccess

	// CRUDAccess Verbs.
	CRUDAccess = GetAccess | ListAccess | DeleteAccess | ViewAccess | EditAccess

	// AllVerbsAccess super powers.
	AllVerbsAccess = CRUDAccess | NamespaceAccess
)

type (
	// RowEvent represents a call for action after a resource reconciliation.
	// Tracks whether a resource got added, deleted or updated.
	RowEvent struct {
		Action watch.EventType
		Fields Row
		Deltas Row
	}

	// RowEvents tracks resource update events.
	RowEvents map[string]*RowEvent

	// TypeMeta represents resource type meta data.
	TypeMeta struct {
		Name       string
		Namespaced bool
		Group      string
		Version    string
		Kind       string
		Singular   string
		Plural     string
		ShortNames []string
	}

	// TableData tracks a K8s resource for tabular display.
	TableData struct {
		Header    Row
		Rows      RowEvents
		NumCols   map[string]bool
		Namespace string
	}

	// List protocol to display and update a collection of resources
	List interface {
		Data() TableData
		Resource() Resource
		Namespaced() bool
		AllNamespaces() bool
		GetNamespace() string
		SetNamespace(string)
		Reconcile(informer *wa.Informer, path *string) error
		GetName() string
		Access(flag int) bool
		GetAccess() int
		SetAccess(int)
		SetFieldSelector(string)
		SetLabelSelector(string)
		HasSelectors() bool
	}

	// Columnar tracks resources that can be diplayed in a tabular fashion.
	Columnar interface {
		Header(ns string) Row
		Fields(ns string) Row
		ExtFields() (TypeMeta, error)
		Name() string
		SetPodMetrics(*mv1beta1.PodMetrics)
		SetNodeMetrics(*mv1beta1.NodeMetrics)
	}

	// Columnars a collection of columnars.
	Columnars []Columnar

	// Row represents a collection of string fields.
	Row []string

	// Rows represents a collection of rows.
	Rows []Row

	// Resource represents a tabular Kubernetes resource.
	Resource interface {
		New(interface{}) Columnar
		Get(path string) (Columnar, error)
		List(ns string, opts metav1.ListOptions) (Columnars, error)
		Delete(path string, cascade, force bool) error
		Describe(gvr, pa string) (string, error)
		Marshal(pa string) (string, error)
		Header(ns string) Row
		NumCols(ns string) map[string]bool
	}

	list struct {
		namespace, name string
		verbs           int
		resource        Resource
		cache           RowEvents
		fieldSelector   string
		labelSelector   string
	}
)

func newRowEvent(a watch.EventType, f, d Row) *RowEvent {
	return &RowEvent{Action: a, Fields: f, Deltas: d}
}

// NewList returns a new resource list.
func NewList(ns, name string, res Resource, verbs int) *list {
	return &list{
		namespace: ns,
		name:      name,
		verbs:     verbs,
		resource:  res,
		cache:     RowEvents{},
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
	l.cache = RowEvents{}
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
func (l *list) Data() TableData {
	return TableData{
		Header:    l.resource.Header(l.namespace),
		Rows:      l.cache,
		NumCols:   l.resource.NumCols(l.namespace),
		Namespace: l.namespace,
	}
}

func (l *list) load(informer *wa.Informer, ns string) (Columnars, error) {
	rr, err := informer.List(l.name, ns, metav1.ListOptions{
		FieldSelector: l.fieldSelector,
		LabelSelector: l.labelSelector,
	})
	if err != nil {
		return nil, err
	}

	items := make(Columnars, 0, len(rr))
	for _, r := range rr {
		res, err := l.fetchResource(informer, r, ns)
		if err != nil {
			return nil, err
		}
		items = append(items, res)
	}

	return items, nil
}

func (l *list) fetchResource(informer *wa.Informer, r interface{}, ns string) (Columnar, error) {
	res := l.resource.New(r)

	switch o := r.(type) {
	case *v1.Node:
		fqn := MetaFQN(o.ObjectMeta)
		if nmx, err := informer.Get(wa.NodeMXIndex, fqn, metav1.GetOptions{}); err != nil {
			return res, err
		} else {
			res.SetNodeMetrics(nmx.(*mv1beta1.NodeMetrics))
		}
	case *v1.Pod:
		fqn := MetaFQN(o.ObjectMeta)
		if pmx, err := informer.Get(wa.PodMXIndex, fqn, metav1.GetOptions{}); err != nil {
			return res, err
		} else {
			res.SetPodMetrics(pmx.(*mv1beta1.PodMetrics))
		}
	case v1.Container:
		if pmx, err := informer.Get(wa.PodMXIndex, ns, metav1.GetOptions{}); err != nil {
			return res, err
		} else {
			res.SetPodMetrics(pmx.(*mv1beta1.PodMetrics))
		}
	default:
		return res, fmt.Errorf("No informer matched %s:%s", l.name, ns)
	}

	return res, nil
}

// Reconcile previous vs current state and emits delta events.
func (l *list) Reconcile(informer *wa.Informer, path *string) error {
	ns := l.namespace
	if path != nil {
		ns = *path
	}

	if items, err := l.load(informer, ns); err == nil {
		l.update(items)
		return nil
	}

	opts := metav1.ListOptions{
		LabelSelector: l.labelSelector,
		FieldSelector: l.fieldSelector,
	}
	if items, err := l.resource.List(l.namespace, opts); err == nil {
		l.update(items)
	} else {
		return err
	}

	return nil
}

func (l *list) update(items Columnars) {
	first := len(l.cache) == 0
	kk := make([]string, 0, len(items))
	for _, i := range items {
		kk = append(kk, i.Name())
		ff := i.Fields(l.namespace)
		if first {
			l.cache[i.Name()] = newRowEvent(New, ff, make(Row, len(ff)))
			continue
		}
		dd := make(Row, len(ff))
		a := watch.Added
		if evt, ok := l.cache[i.Name()]; ok {
			a = computeDeltas(evt, ff[:len(ff)-1], dd)
		}
		l.cache[i.Name()] = newRowEvent(a, ff, dd)
	}

	if first {
		return
	}
	l.ensureDeletes(kk)
}

// EnsureDeletes delete items in cache that are no longer valid.
func (l *list) ensureDeletes(kk []string) {
	for k := range l.cache {
		var found bool
		for i, key := range kk {
			if k == key {
				found = true
				kk = append(kk[:i], kk[i+1:]...)
				break
			}
		}
		if !found {
			delete(l.cache, k)
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

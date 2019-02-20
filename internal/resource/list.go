package resource

import (
	"reflect"
	"sort"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
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
	// SortFn provides for sorting items in  list.
	SortFn func([]string)

	// RowEvent represents a call for action after a resource reconciliation.
	// Tracks whether a resource got added, deleted or updated.
	RowEvent struct {
		Action watch.EventType
		Fields Row
		Deltas Row
	}

	// RowEvents tracks resource update events.
	RowEvents map[string]*RowEvent

	// Properties a collection of extra properties on a K8s resource.
	Properties map[string]interface{}

	// TableData tracks a K8s resource for tabular display.
	TableData struct {
		Header    Row
		Rows      RowEvents
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
		Reconcile() error
		Describe(pa string) (Properties, error)
		GetName() string
		Access(flag int) bool
		HasXRay() bool
		SortFn() SortFn
	}

	// Columnar tracks resources that can be diplayed in a tabular fashion.
	Columnar interface {
		Header(ns string) Row
		Fields(ns string) Row
		ExtFields() Properties
		Name() string
	}

	// Row represents a collection of string fields.
	Row []string

	// Columnars a collection of columnars.
	Columnars []Columnar

	// MxColumnar tracks resource metrics.
	MxColumnar interface {
		Columnar
		Metrics() k8s.Metric
		SetMetrics(k8s.Metric)
	}

	// Resource tracks generic Kubernetes resources.
	Resource interface {
		NewInstance(interface{}) Columnar
		Get(path string) (Columnar, error)
		List(ns string) (Columnars, error)
		Delete(path string) error
		Describe(kind, pa string) (string, error)
		Marshal(pa string) (string, error)
		Header(ns string) Row
	}

	list struct {
		namespace, name string
		verbs           int
		xray            bool
		api             Resource
		cache           RowEvents
		sortFn          func([]string)
	}
)

func newRowEvent(a watch.EventType, f, d Row) *RowEvent {
	return &RowEvent{Action: a, Fields: f, Deltas: d}
}

func newList(ns, name string, api Resource, v int) *list {
	return &list{
		namespace: ns,
		name:      name,
		verbs:     v,
		api:       api,
		cache:     RowEvents{},
	}
}

func (l *list) SortFn() SortFn {
	if l.sortFn == nil {
		return sort.Strings
	}
	return l.sortFn
}

func (l *list) HasXRay() bool {
	return l.xray
}

// Access check access control on a given resource.
func (l *list) Access(f int) bool {
	return l.verbs&f == f
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
	return l.api
}

// Cache tracks previous resource state.
func (l *list) Data() TableData {
	return TableData{
		Header:    l.api.Header(l.namespace),
		Rows:      l.cache,
		Namespace: l.namespace,
	}
}

func (l *list) Describe(pa string) (Properties, error) {
	var p Properties
	i, err := l.api.Get(pa)
	if err != nil {
		return p, err
	}

	return i.ExtFields(), nil
}

// Reconcile previous vs current state and emits delta events.
func (l *list) Reconcile() error {
	var (
		items Columnars
		err   error
	)

	log.Debugf("Fetching list for resource `%s` in ns `%s`", l.name, l.namespace)
	if items, err = l.api.List(l.namespace); err != nil {
		return err
	}

	if len(l.cache) == 0 {
		for _, i := range items {
			ff := i.Fields(l.namespace)
			l.cache[i.Name()] = newRowEvent(New, ff, make(Row, len(ff)))
		}
		return nil
	}

	kk := make([]string, 0, len(items))
	for _, i := range items {
		a := watch.Added
		ff := i.Fields(l.namespace)
		dd := make(Row, len(ff))
		kk = append(kk, i.Name())
		if evt, ok := l.cache[i.Name()]; ok {
			f1, f2 := evt.Fields[:len(evt.Fields)-1], ff[:len(ff)-1]
			a = Unchanged
			if !reflect.DeepEqual(f1, f2) {
				for i, f := range f1 {
					if f != f2[i] {
						dd[i] = f
					}
				}
				a = watch.Modified
			}
		}
		l.cache[i.Name()] = newRowEvent(a, ff, dd)
	}

	// Check for deletions!
	for k := range l.cache {
		var found bool
		for _, key := range kk {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			delete(l.cache, k)
		}
	}
	return nil
}

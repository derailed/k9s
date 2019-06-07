package resource

import (
	"fmt"
	"reflect"

	wa "github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

	// Properties a collection of extra properties on a K8s resource.
	Properties map[string]interface{}

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
		ExtFields() Properties
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
		List(ns string) (Columnars, error)
		Delete(path string, cascade, force bool) error
		Describe(kind, pa string, flags *genericclioptions.ConfigFlags) (string, error)
		Marshal(pa string) (string, error)
		Header(ns string) Row
		NumCols(ns string) map[string]bool
		SetFieldSelector(string)
		SetLabelSelector(string)
		GetFieldSelector() string
		GetLabelSelector() string
		HasSelectors() bool
	}

	list struct {
		namespace, name string
		verbs           int
		resource        Resource
		cache           RowEvents
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
	return l.resource.HasSelectors()
}

// SetFieldSelector narrows down resource query given fields selection.
func (l *list) SetFieldSelector(s string) {
	l.resource.SetFieldSelector(s)
}

// SetLabelSelector narrows down resource query via labels selections.
func (l *list) SetLabelSelector(s string) {
	l.resource.SetLabelSelector(s)
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
	if l.namespace == NotNamespaced {
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

func metaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return fqn(m.Namespace, m.Name)
}

func (l *list) fetchFromStore(informer *wa.Informer, ns string) (Columnars, error) {
	rr, err := informer.List(l.name, ns, metav1.ListOptions{
		FieldSelector: l.resource.GetFieldSelector(),
		LabelSelector: l.resource.GetLabelSelector(),
	})
	if err != nil {
		return nil, err
	}

	items := make(Columnars, 0, len(rr))
	opts := metav1.GetOptions{}
	for _, r := range rr {
		var (
			fqn string
			res Columnar
		)
		switch o := r.(type) {
		case *v1.Node:
			fqn = metaFQN(o.ObjectMeta)
			res = l.resource.New(r)
			nmx, err := informer.Get(wa.NodeMXIndex, fqn, opts)
			if err != nil {
				log.Warn().Err(err).Msg("NodeMetrics")
			}
			if mx, ok := nmx.(*mv1beta1.NodeMetrics); ok {
				res.SetNodeMetrics(mx)
			}
		case *v1.Pod:
			fqn = metaFQN(o.ObjectMeta)
			res = l.resource.New(r)
			pmx, err := informer.Get(wa.PodMXIndex, fqn, opts)
			if err != nil {
				log.Warn().Err(err).Msgf("PodMetrics %s", fqn)
			}
			if mx, ok := pmx.(*mv1beta1.PodMetrics); ok {
				res.SetPodMetrics(mx)
			}
		case v1.Container:
			fqn = ns
			res = l.resource.New(r)
			pmx, err := informer.Get(wa.PodMXIndex, fqn, opts)
			if err != nil {
				log.Warn().Err(err).Msgf("PodMetrics<container> %s", fqn)
			}
			if mx, ok := pmx.(*mv1beta1.PodMetrics); ok {
				res.SetPodMetrics(mx)
			}
		default:
			return items, fmt.Errorf("No informer matched %s:%s", l.name, ns)
		}
		items = append(items, res)
	}

	return items, nil
}

// Reconcile previous vs current state and emits delta events.
func (l *list) Reconcile(informer *wa.Informer, path *string) error {
	ns := l.namespace
	if path != nil {
		ns = *path
	}

	var items Columnars
	if rr, err := l.fetchFromStore(informer, ns); err == nil {
		items = rr
	} else {
		items, err = l.resource.List(l.namespace)
		if err != nil {
			return err
		}
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

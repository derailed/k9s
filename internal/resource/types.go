package resource

import (
	"context"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
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

// Connection represents an apiserver connection.
type Connection k8s.Connection

// TypeName captures resource names.
type TypeName struct {
	Singular   string
	Plural     string
	ShortNames []string
}

// TypeMeta represents resource type meta data.
type TypeMeta struct {
	TypeName

	Name       string
	Namespaced bool
	Group      string
	Version    string
	Kind       string
}

// List protocol to display and update a collection of resources
type List interface {
	Data() render.TableData
	Resource() Resource
	Namespaced() bool
	AllNamespaces() bool
	GetNamespace() string
	SetNamespace(string)
	Reconcile(ctx context.Context, gvr string) error
	GetName() string
	Access(flag int) bool
	GetAccess() int
	SetAccess(int)
	SetFieldSelector(string)
	SetLabelSelector(string)
	HasSelectors() bool
}

// Columnar tracks resources that can be diplayed in a tabular fashion.
type Columnar interface {
	Header(ns string) Row
	Fields(ns string) Row
	ExtFields() (TypeMeta, error)
	Name() string
	SetPodMetrics(*mv1beta1.PodMetrics)
	SetNodeMetrics(*mv1beta1.NodeMetrics)
}

// Columnars a collection of columnars.
type Columnars []Columnar

// Row represents a collection of string fields.
type Row []string

// Rows represents a collection of rows.
type Rows []Row

// Resource represents a tabular Kubernetes resource.
type Resource interface {
	New(interface{}) (Columnar, error)
	Get(path string) (Columnar, error)
	// BOZO!!
	// List(ctx context.Context, ns string) (Columnars, error)
	Delete(path string, cascade, force bool) error
	Describe(gvr, pa string) (string, error)
	Marshal(pa string) (string, error)
	Header(ns string) Row
	NumCols(ns string) map[string]bool
}

// Cruder represents a CRUD operation on a resource.
type Cruder interface {
	// Get retrieves a resource instance.
	Get(ns string, name string) (interface{}, error)

	// BOZO!!
	// List retrieves a resource collection.
	// List(ctx context.Context, ns string) (k8s.Collection, error)

	// Delete remove a resource.
	Delete(ns string, name string, cascade, force bool) error
}

// Scalable represents a scalable resource.
type Scalable interface {
	// Scale scales a resource to a given number of replicas.
	Scale(ns string, name string, replicas int32) error
}

// Restartable represents a restartable resource.
type Restartable interface {
	// Restart performs a rollout restart on a resource
	Restart(ns string, name string) error
}

// Factory creates new tabular resources.
type Factory interface {
	New(interface{}) (Columnar, error)
}

// Containers represents a resource that supports containers.
type Containers interface {
	Containers(path string, includeInit bool) ([]string, error)
}

// Tailable represents a resource with tailable logs.
type Tailable interface {
	Logs(ctx context.Context, c chan<- string, opts LogOptions) error
}

// TailableResource is a resource that have tailable logs.
type TailableResource interface {
	Resource
	Tailable
}

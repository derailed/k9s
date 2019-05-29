package resource

import (
	"bytes"
	"path"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"k8s.io/kubernetes/pkg/kubectl/describe"
	versioned "k8s.io/kubernetes/pkg/kubectl/describe/versioned"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Cruder represent a crudable Kubernetes resource.
	Cruder interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (k8s.Collection, error)
		Delete(ns string, name string) error
		SetLabelSelector(string)
		SetFieldSelector(string)
		GetLabelSelector() string
		GetFieldSelector() string
		HasSelectors() bool
	}

	// Connection represents a Kubenetes apiserver connection.
	Connection k8s.Connection

	// Factory creates new tabular resources.
	Factory interface {
		New(interface{}) Columnar
	}

	// Base resource.
	Base struct {
		Factory

		Connection Connection
		path       string
		Resource   Cruder
	}
)

// NewBase returns a new base
func NewBase(c Connection, r Cruder) *Base {
	return &Base{Connection: c, Resource: r}
}

// HasSelectors returns true if field or label selectors are set.
func (b *Base) HasSelectors() bool {
	return b.Resource.HasSelectors()
}

// SetPodMetrics attach pod metrics to resource.
func (b *Base) SetPodMetrics(*mv1beta1.PodMetrics) {}

// SetNodeMetrics attach node metrics to resource.
func (b *Base) SetNodeMetrics(*mv1beta1.NodeMetrics) {}

// SetFieldSelector refines query results via selector.
func (b *Base) SetFieldSelector(s string) {
	b.Resource.SetFieldSelector(s)
}

// SetLabelSelector refines query results via labels.
func (b *Base) SetLabelSelector(s string) {
	b.Resource.SetLabelSelector(s)
}

// GetFieldSelector returns field selector.
func (b *Base) GetFieldSelector() string {
	return b.Resource.GetFieldSelector()
}

// GetLabelSelector returns label selector.
func (b *Base) GetLabelSelector() string {
	return b.Resource.GetLabelSelector()
}

// Name returns the resource namespaced name.
func (b *Base) Name() string {
	return b.path
}

// NumCols designates if column is numerical.
func (*Base) NumCols(n string) map[string]bool {
	return map[string]bool{}
}

// ExtFields returns extended fields in relation to headers.
func (*Base) ExtFields() Properties {
	return Properties{}
}

// Get a resource by name
func (b *Base) Get(path string) (Columnar, error) {
	ns, n := namespaced(path)
	i, err := b.Resource.Get(ns, n)
	if err != nil {
		return nil, err
	}

	return b.New(i), nil
}

// List all resources
func (b *Base) List(ns string) (Columnars, error) {
	ii, err := b.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	cc := make(Columnars, 0, len(ii))
	for i := 0; i < len(ii); i++ {
		cc = append(cc, b.New(ii[i]))
	}

	return cc, nil
}

// Describe a given resource.
func (b *Base) Describe(kind, pa string, flags *genericclioptions.ConfigFlags) (string, error) {
	ns, n := namespaced(pa)

	mapping, err := k8s.RestMapping.Find(kind)
	if err != nil {
		log.Debug().Msgf("Unable to find mapper for %s %s", kind, pa)
		return "", err
	}

	d, err := versioned.Describer(flags, mapping)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find describer for %#v", mapping)
		return "", err
	}

	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}

// Delete a resource by name.
func (b *Base) Delete(path string) error {
	ns, n := namespaced(path)

	return b.Resource.Delete(ns, n)
}

func (*Base) namespacedName(m metav1.ObjectMeta) string {
	return path.Join(m.Namespace, m.Name)
}

func (*Base) marshalObject(o runtime.Object) (string, error) {
	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	err := p.PrintObj(o, &buff)
	if err != nil {
		log.Error().Msgf("Marshal Error %v", err)
		return "", err
	}

	return buff.String(), nil
}

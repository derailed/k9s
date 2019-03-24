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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/kubectl/describe"
	versioned "k8s.io/kubernetes/pkg/kubectl/describe/versioned"
	v "k8s.io/metrics/pkg/client/clientset/versioned"
)

type (
	// Cruder represent a crudable Kubernetes resource.
	Cruder interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (k8s.Collection, error)
		Delete(ns string, name string) error
	}

	// Connection represents a Kubenetes apiserver connection.
	Connection interface {
		Config() *k8s.Config
		DialOrDie() kubernetes.Interface
		SwitchContextOrDie(ctx string)
		NSDialOrDie() dynamic.NamespaceableResourceInterface
		RestConfigOrDie() *restclient.Config
		MXDial() (*v.Clientset, error)
		DynDialOrDie() dynamic.Interface
		HasMetrics() bool
		IsNamespaced(n string) bool
		SupportsResource(group string) bool
	}

	// Factory creates new tabular resources.
	Factory interface {
		New(interface{}) Columnar
	}

	// Base resource.
	Base struct {
		Factory

		connection Connection
		path       string
		resource   Cruder
	}
)

// Name returns the resource namespaced name.
func (b *Base) Name() string {
	return b.path
}

// ExtFields returns extended fields in relation to headers.
func (*Base) ExtFields() Properties {
	return Properties{}
}

// Get a resource by name
func (b *Base) Get(path string) (Columnar, error) {
	ns, n := namespaced(path)
	i, err := b.resource.Get(ns, n)
	if err != nil {
		return nil, err
	}

	return b.New(i), nil
}

// List all resources
func (b *Base) List(ns string) (Columnars, error) {
	ii, err := b.resource.List(ns)
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
		return "", err
	}

	d, err := versioned.Describer(flags, mapping)
	if err != nil {
		return "", err
	}
	opts := describe.DescriberSettings{
		ShowEvents: true,
	}

	return d.Describe(ns, n, opts)
}

// Delete a resource by name.
func (b *Base) Delete(path string) error {
	ns, n := namespaced(path)

	return b.resource.Delete(ns, n)
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

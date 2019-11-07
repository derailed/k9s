package resource

import (
	"bytes"
	"context"
	"errors"
	"path"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericprinters "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/describe/versioned"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Cruder represents a crudable Kubernetes resource.
	Cruder interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string, opts metav1.ListOptions) (k8s.Collection, error)
		Delete(ns string, name string, cascade, force bool) error
	}

	// Scalable represents a scalable Kubernetes resource.
	Scalable interface {
		Scale(ns string, name string, replicas int32) error
	}

	// Restartable represents a rollout restartable Kubernetes resource.
	Restartable interface {
		Restart(ns string, name string) error
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

// SetPodMetrics attach pod metrics to resource.
func (b *Base) SetPodMetrics(*mv1beta1.PodMetrics) {}

// SetNodeMetrics attach node metrics to resource.
func (b *Base) SetNodeMetrics(*mv1beta1.NodeMetrics) {}

// Name returns the resource namespaced name.
func (b *Base) Name() string {
	return b.path
}

// NumCols designates if column is numerical.
func (*Base) NumCols(n string) map[string]bool {
	return map[string]bool{}
}

// ExtFields returns extended fields in relation to headers.
func (*Base) ExtFields() (TypeMeta, error) {
	return TypeMeta{}, errors.New("Base does not have extended fields.")
}

// Get a resource by name
func (b *Base) Get(path string) (Columnar, error) {
	ns, n := Namespaced(path)
	i, err := b.Resource.Get(ns, n)
	if err != nil {
		return nil, err
	}

	return b.New(i), nil
}

// List all resources
func (b *Base) List(ns string, opts metav1.ListOptions) (Columnars, error) {
	ii, err := b.Resource.List(ns, opts)
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
func (b *Base) Describe(gvr, pa string) (string, error) {
	mapper := k8s.RestMapper{Connection: b.Connection}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		log.Error().Err(err).Msgf("No REST mapper for resource %s", gvr)
		return "", err
	}

	GVR := k8s.GVR(gvr)
	gvk, err := m.KindFor(GVR.AsGVR())
	if err != nil {
		log.Error().Err(err).Msgf("No GVK for resource %s", gvr)
		return "", err
	}

	mapping, err := mapper.ResourceFor(GVR.ResName(), gvk.Kind)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find mapper for %s %s", gvr, pa)
		return "", err
	}
	ns, n := Namespaced(pa)
	d, err := versioned.Describer(b.Connection.Config().Flags(), mapping)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find describer for %#v", mapping)
		return "", err
	}

	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}

// Delete a resource by name.
func (b *Base) Delete(path string, cascade, force bool) error {
	ns, n := Namespaced(path)

	return b.Resource.Delete(ns, n, cascade, force)
}

func (*Base) namespacedName(m metav1.ObjectMeta) string {
	return path.Join(m.Namespace, m.Name)
}

func (*Base) marshalObject(o runtime.Object) (string, error) {
	var (
		buff bytes.Buffer
		p    genericprinters.YAMLPrinter
	)
	err := p.PrintObj(o, &buff)
	if err != nil {
		log.Error().Msgf("Marshal Error %v", err)
		return "", err
	}

	return buff.String(), nil
}

func (b *Base) podLogs(ctx context.Context, c chan<- string, sel map[string]string, opts LogOptions) error {
	i := ctx.Value(IKey("informer")).(*watch.Informer)
	pods, err := i.List(watch.PodIndex, opts.Namespace, metav1.ListOptions{
		LabelSelector: toSelector(sel),
	})
	if err != nil {
		return err
	}

	if len(pods) > 1 {
		opts.MultiPods = true
	}
	pr := NewPod(b.Connection)
	for _, p := range pods {
		po := p.(*v1.Pod)
		if po.Status.Phase == v1.PodRunning {
			opts.Namespace, opts.Name = po.Namespace, po.Name
			if err := pr.PodLogs(ctx, c, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

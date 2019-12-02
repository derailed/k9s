package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Resource represents a Kubernetes Resource
type Resource struct {
	*base
	Connection

	gvr GVR
}

// NewResource returns a new Resource.
func NewResource(c Connection, gvr GVR) *Resource {
	return &Resource{base: &base{}, Connection: c, gvr: gvr}
}

// GetInfo returns info about apigroup.
func (r *Resource) GetInfo() GVR {
	return r.gvr
}

func (r *Resource) nsRes() dynamic.NamespaceableResourceInterface {
	return r.DynDialOrDie().Resource(r.gvr.AsGVR())
}

// Get a Resource.
func (r *Resource) Get(ns, n string) (interface{}, error) {
	return r.nsRes().Namespace(ns).Get(n, metav1.GetOptions{})
}

// List all Resources in a given namespace.
func (r *Resource) List(ns string, opts metav1.ListOptions) (Collection, error) {
	obj, err := r.listAll(ns, r.gvr.ToR())
	if err != nil {
		return nil, err
	}
	return Collection{obj.(*metav1beta1.Table)}, nil
}

// Delete a Resource.
func (r *Resource) Delete(ns, n string, cascade, force bool) error {
	return r.nsRes().Namespace(ns).Delete(n, nil)
}

// ----------------------------------------------------------------------------
// Helpers...

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

func (r *Resource) listAll(ns, n string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := r.codec()

	c, err := r.getClient()
	if err != nil {
		return nil, err
	}

	return c.Get().
		SetHeader("Accept", a).
		Namespace(ns).
		Resource(n).
		VersionedParams(&metav1beta1.TableOptions{}, codec).
		Do().Get()
}

func (r *Resource) getClient() (*rest.RESTClient, error) {
	crConfig := r.RestConfigOrDie()
	gv := r.gvr.AsGV()
	crConfig.GroupVersion = &gv
	crConfig.APIPath = "/apis"
	if len(r.gvr.ToG()) == 0 {
		crConfig.APIPath = "/api"
	}
	codec, _ := r.codec()
	crConfig.NegotiatedSerializer = codec.WithoutConversion()

	crRestClient, err := rest.RESTClientFor(crConfig)
	if err != nil {
		return nil, err
	}
	return crRestClient, nil
}

func (r *Resource) codec() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := r.gvr.AsGV()
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}

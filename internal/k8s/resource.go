package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Resource represents a Kubernetes Resource
type Resource struct {
	Connection

	group, version, name string
}

// NewResource returns a new Resource.
func NewResource(c Connection, group, version, name string) Cruder {
	return &Resource{Connection: c, group: group, version: version, name: name}
}

// GetInfo returns info about apigroup.
func (r *Resource) GetInfo() (string, string, string) {
	return r.group, r.version, r.name
}

func (r *Resource) base() dynamic.NamespaceableResourceInterface {
	g := schema.GroupVersionResource{
		Group:    r.group,
		Version:  r.version,
		Resource: r.name,
	}
	return r.DynDialOrDie().Resource(g)
}

// Get a Resource.
func (r *Resource) Get(ns, n string) (interface{}, error) {
	return r.base().Namespace(ns).Get(n, metav1.GetOptions{})
}

// List all Resources in a given namespace.
func (r *Resource) List(ns string) (Collection, error) {
	obj, err := r.listAll(ns, r.name)
	if err != nil {
		return nil, err
	}
	return Collection{obj.(*metav1beta1.Table)}, nil
}

// Delete a Resource.
func (r *Resource) Delete(ns, n string) error {
	return r.base().Namespace(ns).Delete(n, nil)
}

// ----------------------------------------------------------------------------
// Helpers...

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

func (r *Resource) listAll(ns, n string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := r.codecs()

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

func (r *Resource) getOne(ns, n string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := r.codecs()

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
	crConfig.GroupVersion = &schema.GroupVersion{Group: r.group, Version: r.version}
	crConfig.APIPath = "/apis"
	if len(r.group) == 0 {
		crConfig.APIPath = "/api"
	}
	codecs, _ := r.codecs()
	crConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: codecs}

	crRestClient, err := rest.RESTClientFor(crConfig)
	if err != nil {
		return nil, err
	}
	return crRestClient, nil
}

func (r *Resource) codecs() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: r.group, Version: r.version}
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}

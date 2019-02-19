package k8s

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	group, version, name string
}

// NewResource returns a new Resource.
func NewResource(group, version, name string) Res {
	return &Resource{group: group, version: version, name: name}
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
	return conn.dynDialOrDie().Resource(g)
}

// Get a Resource.
func (r *Resource) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return r.base().Namespace(ns).Get(n, opts)
}

// List all Resources in a given namespace
func (r *Resource) List(ns string) (Collection, error) {
	obj, err := r.listAll(ns, r.name)
	if err != nil {
		return Collection{}, err
	}

	return Collection{obj.(*metav1beta1.Table)}, nil
}

// Delete a Resource
func (r *Resource) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return r.base().Namespace(ns).Delete(n, &opts)
}

// Helpers...

func (r *Resource) getClient() *rest.RESTClient {
	gv := schema.GroupVersion{Group: r.group, Version: r.version}
	codecs, _ := r.codecs()
	crConfig := *conn.restConfigOrDie()
	crConfig.GroupVersion = &gv
	crConfig.APIPath = "/apis"
	if len(r.group) == 0 {
		crConfig.APIPath = "/api"
	}
	crConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: codecs}
	crRestClient, err := rest.RESTClientFor(&crConfig)
	if err != nil {
		log.Fatal(err)
	}
	return crRestClient
}

func (r *Resource) codecs() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: r.group, Version: r.version}
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	codecs := serializer.NewCodecFactory(scheme)
	return codecs, runtime.NewParameterCodec(scheme)
}

func (r *Resource) listAll(ns, n string) (runtime.Object, error) {
	group := metav1beta1.GroupName
	version := metav1beta1.SchemeGroupVersion.Version
	a := fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)

	_, codec := r.codecs()
	return r.getClient().Get().
		SetHeader("Accept", a).
		Namespace(ns).
		Resource(n).
		VersionedParams(&metav1beta1.TableOptions{}, codec).
		Do().
		Get()
}

func (r *Resource) getOne(ns, n string) (runtime.Object, error) {
	group := metav1beta1.GroupName
	version := metav1beta1.SchemeGroupVersion.Version
	a := fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)

	_, codec := r.codecs()
	return r.getClient().Get().
		SetHeader("Accept", a).
		Namespace(ns).
		Resource(n).
		VersionedParams(&metav1beta1.TableOptions{}, codec).
		Do().
		Get()
}

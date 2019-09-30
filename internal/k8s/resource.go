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

// CustomResource represents a Kubernetes CustomResource
type CustomResource struct {
	*Resource
	Connection
}

// NewCustomResource returns a new CustomResource.
func NewCustomResource(c Connection, gvr GVR) *CustomResource {
	return &CustomResource{Resource: &Resource{gvr: gvr}, Connection: c}
}

func (cr *CustomResource) nsRes() dynamic.NamespaceableResourceInterface {
	g := schema.GroupVersionResource{
		Group:    cr.gvr.Group,
		Version:  cr.gvr.Version,
		Resource: cr.gvr.Resource,
	}
	return cr.DynDialOrDie().Resource(g)
}

// Get a CustomResource.
func (cr *CustomResource) Get(ns, n string) (interface{}, error) {
	return cr.nsRes().Namespace(ns).Get(n, metav1.GetOptions{})
}

// List all Resources in a given namespace.
func (cr *CustomResource) List(ns string) (Collection, error) {
	obj, err := cr.listAll(ns, cr.gvr.Resource)
	if err != nil {
		return nil, err
	}
	return Collection{obj.(*metav1beta1.Table)}, nil
}

// Delete a CustomResource.
func (cr *CustomResource) Delete(ns, n string, cascade, force bool) error {
	return cr.nsRes().Namespace(ns).Delete(n, nil)
}

// ----------------------------------------------------------------------------
// Helpers...

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

func (cr *CustomResource) listAll(ns, n string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := cr.codecs()

	c, err := cr.getClient()
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

func (cr *CustomResource) getClient() (*rest.RESTClient, error) {
	crConfig := cr.RestConfigOrDie()
	crConfig.GroupVersion = &schema.GroupVersion{Group: cr.gvr.Group, Version: cr.gvr.Version}
	crConfig.APIPath = "/apis"
	if len(cr.gvr.Group) == 0 {
		crConfig.APIPath = "/api"
	}
	codecs, _ := cr.codecs()
	crConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: codecs}

	crRestClient, err := rest.RESTClientFor(crConfig)
	if err != nil {
		return nil, err
	}
	return crRestClient, nil
}

func (cr *CustomResource) codecs() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: cr.gvr.Group, Version: cr.gvr.Version}
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}

package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

// Generic represents a generic model.
type Generic struct {
	Resource

	table *metav1beta1.Table
}

// List returns a collection of node resources.
func (g *Generic) Get(ctx context.Context, path string) (runtime.Object, error) {
	var opts metav1.GetOptions

	ns, n := client.Namespaced(path)
	gvr := client.NewGVR(g.gvr)
	req := g.factory.Client().DynDialOrDie().Resource(gvr.AsGVR())
	if ns == "" {
		return req.Get(n, opts)
	}

	return req.Namespace(ns).Get(n, opts)
}

// List returns a collection of node resources.
func (g *Generic) List(ctx context.Context) ([]runtime.Object, error) {
	// Ensures the factory is tracking this resource
	_, err := g.factory.CanForResource(g.namespace, g.gvr, []string{"list"})
	if err != nil {
		return nil, err
	}

	gvr := client.NewGVR(g.gvr)
	fcodec, codec := g.codec(gvr.AsGV())

	c, err := g.client(fcodec, gvr)
	if err != nil {
		return nil, err
	}

	ns := g.namespace
	if g.namespace == render.ClusterScope {
		ns = render.AllNamespaces
	}

	log.Debug().Msgf("GENERIC LIST %q:%q", g.namespace, g.gvr)
	o, err := c.Get().
		SetHeader("Accept", fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)).
		Resource(gvr.ToR()).
		VersionedParams(&metav1beta1.TableOptions{}, codec).
		Namespace(ns).
		Do().
		Get()
	if err != nil {
		return nil, err
	}
	table, ok := o.(*metav1beta1.Table)
	if !ok {
		return nil, fmt.Errorf("expecting table but got %T", o)
	}
	g.table = table
	res := make([]runtime.Object, len(g.table.Rows))
	for i := range g.table.Rows {
		res[i] = RowRes{&g.table.Rows[i]}
	}

	return res, err
}

// Hydrate returns nodes as rows.
func (g *Generic) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	gr, ok := re.(*render.Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer for %s but got %T", g.gvr, re)
	}
	gr.SetTable(g.table)
	for i, o := range oo {
		res, ok := o.(RowRes)
		if !ok {
			return fmt.Errorf("expecting RowRes but got %#v", o)
		}
		if err := gr.Render(res.TableRow, g.namespace, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func (g *Generic) client(codec serializer.CodecFactory, gvr client.GVR) (*rest.RESTClient, error) {
	crConfig := g.factory.Client().RestConfigOrDie()
	gv := gvr.AsGV()
	crConfig.GroupVersion = &gv
	crConfig.APIPath = "/apis"
	if len(gvr.ToG()) == 0 {
		crConfig.APIPath = "/api"
	}
	crConfig.NegotiatedSerializer = codec.WithoutConversion()

	crRestClient, err := rest.RESTClientFor(crConfig)
	if err != nil {
		return nil, err
	}
	return crRestClient, nil
}

func (r *Resource) codec(gv schema.GroupVersion) (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}

// ----------------------------------------------------------------------------

// RowRes represents a table row.
type RowRes struct {
	*metav1beta1.TableRow
}

// GetObjectKind returns a schema object.
func (r RowRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (r RowRes) DeepCopyObject() runtime.Object {
	return r
}

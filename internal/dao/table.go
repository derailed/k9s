// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

// BOZO!! Figure out how to convert to table def and use factory.

// Table retrieves K8s resources as tabular data.
type Table struct {
	Generic
}

// Get returns a given resource.
func (t *Table) Get(ctx context.Context, path string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := t.codec()

	c, err := t.getClient()
	if err != nil {
		return nil, err
	}
	ns, n := client.Namespaced(path)
	req := c.Get().
		SetHeader("Accept", a).
		Name(n).
		Resource(t.gvr.R()).
		VersionedParams(&metav1beta1.TableOptions{}, codec)
	if ns != client.ClusterScope {
		req = req.Namespace(ns)
	}

	return req.Do(ctx).Get()
}

// List all Resources in a given namespace.
func (t *Table) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labelSel, ok := ctx.Value(internal.KeyLabels).(string)
	if !ok {
		labelSel = ""
	}

	a := fmt.Sprintf(gvFmt, metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName)
	_, codec := t.codec()

	c, err := t.getClient()
	if err != nil {
		return nil, err
	}
	o, err := c.Get().
		SetHeader("Accept", a).
		Namespace(ns).
		Resource(t.gvr.R()).
		VersionedParams(&metav1.ListOptions{LabelSelector: labelSel}, codec).
		Do(ctx).Get()
	if err != nil {
		return nil, err
	}

	return []runtime.Object{o}, nil
}

// ----------------------------------------------------------------------------
// Helpers...

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

func (t *Table) getClient() (*rest.RESTClient, error) {
	cfg, err := t.Client().RestConfig()
	if err != nil {
		return nil, err
	}
	gv := t.gvr.GV()
	cfg.GroupVersion = &gv
	cfg.APIPath = "/apis"
	if t.gvr.G() == "" {
		cfg.APIPath = "/api"
	}
	codec, _ := t.codec()
	cfg.NegotiatedSerializer = codec.WithoutConversion()

	crRestClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}

	return crRestClient, nil
}

func (t *Table) codec() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := t.gvr.GV()
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}

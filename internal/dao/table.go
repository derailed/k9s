// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

// BOZO!! Figure out how to convert to table def and use factory.

var genScheme = runtime.NewScheme()

// Table retrieves K8s resources as tabular data.
type Table struct {
	Generic
}

// Get returns a given resource.
func (t *Table) Get(ctx context.Context, path string) (runtime.Object, error) {
	a := fmt.Sprintf(gvFmt, metav1.SchemeGroupVersion.Version, metav1.GroupName)
	f, p := t.codec()

	c, err := t.getClient(f)
	if err != nil {
		return nil, err
	}
	ns, n := client.Namespaced(path)
	req := c.Get().
		SetHeader("Accept", a).
		Name(n).
		Resource(t.gvr.R()).
		VersionedParams(&metav1.TableOptions{}, p)
	if ns != client.ClusterScope {
		req = req.Namespace(ns)
	}

	return req.Do(ctx).Get()
}

// List all Resources in a given namespace.
func (t *Table) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	defer func(ti time.Time) {
		log.Debug().Msgf(">>>> TABLE-COL [%s] (%s)", t.gvr, time.Since(ti))
	}(time.Now())

	labelSel, _ := ctx.Value(internal.KeyLabels).(string)
	a := fmt.Sprintf(gvFmt, metav1.SchemeGroupVersion.Version, metav1.GroupName)

	fieldSel, _ := ctx.Value(internal.KeyFields).(string)

	f, p := t.codec()

	c, err := t.getClient(f)
	if err != nil {
		return nil, err
	}
	o, err := c.Get().
		SetHeader("Accept", a).
		Namespace(ns).
		Resource(t.gvr.R()).
		VersionedParams(&metav1.ListOptions{
			LabelSelector: labelSel,
			FieldSelector: fieldSel,
		}, p).
		Do(ctx).Get()
	if err != nil {
		return nil, err
	}

	return []runtime.Object{o}, nil
}

// ----------------------------------------------------------------------------
// Helpers...

const gvFmt = "application/json;as=Table;v=%s;g=%s, application/json"

func (t *Table) getClient(f serializer.CodecFactory) (*rest.RESTClient, error) {
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
	cfg.NegotiatedSerializer = f.WithoutConversion()

	crRestClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}

	return crRestClient, nil
}

func (t *Table) codec() (serializer.CodecFactory, runtime.ParameterCodec) {
	var tt metav1.Table
	opts := metav1.TableOptions{IncludeObject: v1.IncludeObject}
	gv := t.gvr.GV()
	metav1.AddToGroupVersion(genScheme, gv)
	genScheme.AddKnownTypes(gv, &tt, &opts)
	genScheme.AddKnownTypes(metav1.SchemeGroupVersion, &tt, &opts)

	return serializer.NewCodecFactory(genScheme), runtime.NewParameterCodec(genScheme)
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s
package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Dynamic struct {
	Generic
}

// Get returns a given resource as a table object.
func (d *Dynamic) Get(ctx context.Context, path string) (runtime.Object, error) {
	oo, err := d.toTable(ctx, path)
	if err != nil || len(oo) == 0 {
		return nil, err
	}

	return oo[0], nil
}

// List returns a collection of resources as one or more table objects.
func (d *Dynamic) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	return d.toTable(ctx, ns+"/")
}

func (d *Dynamic) toTable(ctx context.Context, fqn string) ([]runtime.Object, error) {
	sel := labels.Everything()
	if s, ok := ctx.Value(internal.KeyLabels).(labels.Selector); ok {
		sel = s
	}

	opts := []string{d.gvr.AsResourceName()}
	ns, n := client.Namespaced(fqn)
	if n != "" {
		opts = append(opts, n)
	}
	allNS := client.IsAllNamespaces(ns)
	flags := cmdutil.NewMatchVersionFlags(d.getFactory().Client().Config().Flags())
	f := cmdutil.NewFactory(flags)
	b := f.NewBuilder().
		Unstructured().
		NamespaceParam(ns).DefaultNamespace().AllNamespaces(allNS).
		LabelSelectorParam(sel.String()).
		FieldSelectorParam("").
		RequestChunksOf(0).
		ResourceTypeOrNameArgs(true, opts...).
		ContinueOnError().
		Latest().
		Flatten().
		TransformRequests(d.transformRequests).
		Do()
	if err := b.Err(); err != nil {
		return nil, err
	}

	infos, err := b.Infos()
	if err != nil {
		return nil, err
	}
	oo := make([]runtime.Object, 0, len(infos))
	for _, info := range infos {
		o, err := decodeIntoTable(info.Object, allNS)
		if err != nil {
			return nil, err
		}
		oo = append(oo, o.(*metav1.Table))
	}

	return oo, nil
}

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}

func decodeIntoTable(obj runtime.Object, allNs bool) (runtime.Object, error) {
	event, isEvent := obj.(*metav1.WatchEvent)
	if isEvent {
		obj = event.Object.Object
	}
	if !recognizedTableVersions[obj.GetObjectKind().GroupVersionKind()] {
		return nil, fmt.Errorf("attempt to decode non-Table object: %v", obj.GetObjectKind().GroupVersionKind())
	}

	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("attempt to decode non-Unstructured object")
	}
	var table metav1.Table
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &table); err != nil {
		return nil, err
	}

	if allNs {
		defs := make([]metav1.TableColumnDefinition, 0, len(table.ColumnDefinitions)+1)
		defs = append(defs, metav1.TableColumnDefinition{Name: "Namespace", Type: "string"})
		defs = append(defs, table.ColumnDefinitions...)
		table.ColumnDefinitions = defs
	}

	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			continue
		}
		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			return nil, err
		}
		row.Object.Object = converted
		var m metav1.Object
		if obj := row.Object.Object; obj != nil {
			m, _ = meta.Accessor(obj)
		}
		var ns string
		if m != nil {
			ns = m.GetNamespace()
		}
		if allNs {
			cells := make([]any, 0, len(row.Cells)+1)
			cells = append(cells, ns)
			cells = append(cells, row.Cells...)
			row.Cells = cells
		}
	}

	if isEvent {
		event.Object.Object = &table
		return event, nil
	}

	return &table, nil
}

func (d *Dynamic) transformRequests(req *rest.Request) {
	req.SetHeader("Accept", strings.Join([]string{
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
		"application/json",
	}, ","))

	if d.includeObj {
		req.Param("includeObject", "Object")
	}
}

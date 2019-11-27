package render

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace renders a K8s Namespace to screen.
type Namespace struct{}

// ColorerFunc colors a resource row.
func (Namespace) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header rbw.
func (Namespace) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "AGE"},
	}
}

// Render renders a K8s resource to screen.
func (Namespace) Render(o interface{}, _ string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Namespace, but got %T", o)
	}
	var ns v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ns)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	fields = append(fields,
		ns.Name,
		string(ns.Status.Phase),
		toAge(ns.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(ns.ObjectMeta), fields

	return nil
}

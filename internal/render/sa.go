package render

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceAccount renders a K8s ServiceAccount to screen.
type ServiceAccount struct{}

// ColorerFunc colors a resource row.
func (ServiceAccount) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (ServiceAccount) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "SECRET"},
		Header{Name: "AGE"},
	)
}

// Render renders a K8s resource to screen.
func (ServiceAccount) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ServiceAccount, but got %T", o)
	}
	var s v1.ServiceAccount
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &s)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, s.Namespace)
	}
	fields = append(fields,
		s.Name,
		strconv.Itoa(len(s.Secrets)),
		toAge(s.ObjectMeta.CreationTimestamp),
	)

	r.ID, r.Fields = MetaFQN(s.ObjectMeta), fields

	return nil
}

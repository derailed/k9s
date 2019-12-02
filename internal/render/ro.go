package render

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Role renders a K8s Role to screen.
type Role struct{}

// ColorerFunc colors a resource row.
func (Role) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Role) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "AGE", Decorator: ageDecorator},
	)
}

// Render renders a K8s resource to screen.
func (Role) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Role, but got %T", o)
	}
	var ro rbacv1.Role
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ro)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, ro.Namespace)
	}
	fields = append(fields,
		ro.Name,
		toAge(ro.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(ro.ObjectMeta), fields

	return nil
}

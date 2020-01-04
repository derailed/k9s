package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
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
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (r Role) Render(o interface{}, ns string, row *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Role, but got %T", o)
	}
	var ro rbacv1.Role
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ro)
	if err != nil {
		return err
	}

	row.ID = MetaFQN(ro.ObjectMeta)
	row.Fields = make(Fields, 0, len(r.Header(ns)))
	if client.IsAllNamespaces(ns) {
		row.Fields = append(row.Fields, ro.Namespace)
	}
	row.Fields = append(row.Fields,
		ro.Name,
		toAge(ro.ObjectMeta.CreationTimestamp),
	)

	return nil
}

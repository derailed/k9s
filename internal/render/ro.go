package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Role renders a K8s Role to screen.
type Role struct {
	Base
}

// Header returns a header row.
func (Role) Header(ns string) Header {
	var h Header
	if client.IsAllNamespaces(ns) {
		h = append(h, HeaderColumn{Name: "NAMESPACE"})
	}

	return append(h,
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
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

	row.ID = client.MetaFQN(ro.ObjectMeta)
	row.Fields = make(Fields, 0, len(r.Header(ns)))
	if client.IsAllNamespaces(ns) {
		row.Fields = append(row.Fields, ro.Namespace)
	}
	row.Fields = append(row.Fields,
		ro.Name,
		mapToStr(ro.Labels),
		"",
		toAge(ro.GetCreationTimestamp()),
	)

	return nil
}

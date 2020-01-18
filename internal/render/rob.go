package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RoleBinding renders a K8s RoleBinding to screen.
type RoleBinding struct{}

// ColorerFunc colors a resource row.
func (RoleBinding) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header rbw.
func (RoleBinding) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "ROLE"},
		Header{Name: "KIND"},
		Header{Name: "SUBJECTS"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (r RoleBinding) Render(o interface{}, ns string, row *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected RoleBinding, but got %T", o)
	}
	var rb rbacv1.RoleBinding
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rb)
	if err != nil {
		return err
	}

	kind, ss := renderSubjects(rb.Subjects)

	row.ID = client.MetaFQN(rb.ObjectMeta)
	row.Fields = make(Fields, 0, len(r.Header(ns)))
	if client.IsAllNamespaces(ns) {
		row.Fields = append(row.Fields, rb.Namespace)
	}
	row.Fields = append(row.Fields,
		rb.Name,
		rb.RoleRef.Name,
		kind,
		ss,
		toAge(rb.ObjectMeta.CreationTimestamp),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func renderSubjects(ss []rbacv1.Subject) (kind string, subjects string) {
	if len(ss) == 0 {
		return NAValue, ""
	}

	var tt []string
	for _, s := range ss {
		kind = toSubjectAlias(s.Kind)
		tt = append(tt, s.Name)
	}
	return kind, strings.Join(tt, ",")
}

func toSubjectAlias(s string) string {
	if len(s) == 0 {
		return s
	}

	switch s {
	case rbacv1.UserKind:
		return "USR"
	case rbacv1.GroupKind:
		return "GRP"
	case rbacv1.ServiceAccountKind:
		return "SA"
	default:
		return strings.ToUpper(s)
	}
}

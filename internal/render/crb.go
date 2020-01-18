package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterRoleBinding renders a K8s ClusterRoleBinding to screen.
type ClusterRoleBinding struct{}

// ColorerFunc colors a resource row.
func (ClusterRoleBinding) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header rbw.
func (ClusterRoleBinding) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "CLUSTERROLE"},
		Header{Name: "KIND"},
		Header{Name: "SUBJECTS"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (ClusterRoleBinding) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ClusterRoleBinding, but got %T", o)
	}
	var crb rbacv1.ClusterRoleBinding
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &crb)
	if err != nil {
		return err
	}

	kind, ss := renderSubjects(crb.Subjects)

	r.ID = client.FQN("-", crb.ObjectMeta.Name)
	r.Fields = Fields{
		crb.Name,
		crb.RoleRef.Name,
		kind,
		ss,
		toAge(crb.ObjectMeta.CreationTimestamp),
	}

	return nil
}

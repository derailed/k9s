package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterRole renders a K8s ClusterRole to screen.
type ClusterRole struct{}

// ColorerFunc colors a resource row.
func (ClusterRole) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header rbw.
func (ClusterRole) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (ClusterRole) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expecting clusterrole, but got %T", o)
	}
	var cr rbacv1.ClusterRole
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cr)
	if err != nil {
		return err
	}

	r.ID = client.FQN("-", cr.ObjectMeta.Name)
	r.Fields = Fields{
		cr.Name,
		mapToStr(cr.Labels),
		toAge(cr.ObjectMeta.CreationTimestamp),
	}

	return nil
}

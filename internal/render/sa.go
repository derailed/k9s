package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
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
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "SECRET"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s ServiceAccount) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ServiceAccount, but got %T", o)
	}
	var sa v1.ServiceAccount
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sa.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, sa.Namespace)
	}
	r.Fields = append(r.Fields,
		sa.Name,
		strconv.Itoa(len(sa.Secrets)),
		toAge(sa.ObjectMeta.CreationTimestamp),
	)

	return nil
}

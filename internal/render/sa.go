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
type ServiceAccount struct {
	Base
}

// Header returns a header row.
func (ServiceAccount) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "SECRET"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
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
	r.Fields = Fields{
		sa.Namespace,
		sa.Name,
		strconv.Itoa(len(sa.Secrets)),
		mapToStr(sa.Labels),
		"",
		toAge(sa.ObjectMeta.CreationTimestamp),
	}

	return nil
}

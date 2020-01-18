package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// StorageClass renders a K8s StorageClass to screen.
type StorageClass struct{}

// ColorerFunc colors a resource row.
func (StorageClass) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (StorageClass) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "PROVISIONER"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (StorageClass) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected StorageClass, but got %T", o)
	}
	var sc storagev1.StorageClass
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sc)
	if err != nil {
		return err
	}

	r.ID = client.FQN(client.ClusterScope, sc.ObjectMeta.Name)
	r.Fields = Fields{
		sc.Name,
		string(sc.Provisioner),
		toAge(sc.ObjectMeta.CreationTimestamp),
	}

	return nil
}

package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersistentVolumeClaim renders a K8s PersistentVolumeClaim to screen.
type PersistentVolumeClaim struct{}

// ColorerFunc colors a resource row.
func (PersistentVolumeClaim) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}

		markCol := 2
		if ns != AllNamespaces {
			markCol = 1
		}

		if strings.TrimSpace(r.Row.Fields[markCol]) != "Bound" {
			c = ErrColor
		}

		return c
	}

}

// Header returns a header rbw.
func (PersistentVolumeClaim) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "VOLUME"},
		Header{Name: "CAPACITY"},
		Header{Name: "ACCESS MODES"},
		Header{Name: "STORAGECLASS"},
		Header{Name: "AGE", Decorator: ageDecorator},
	)
}

// Render renders a K8s resource to screen.
func (PersistentVolumeClaim) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected PersistentVolumeClaim, but got %T", o)
	}
	var pvc v1.PersistentVolumeClaim
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &pvc)
	if err != nil {
		return err
	}

	phase := pvc.Status.Phase
	if pvc.ObjectMeta.DeletionTimestamp != nil {
		phase = "Terminating"
	}

	storage := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	var capacity, accessModes string
	if pvc.Spec.VolumeName != "" {
		accessModes = accessMode(pvc.Status.AccessModes)
		storage = pvc.Status.Capacity[v1.ResourceStorage]
		capacity = storage.String()
	}
	class, found := pvc.Annotations[v1.BetaStorageClassAnnotation]
	if !found {
		if pvc.Spec.StorageClassName != nil {
			class = *pvc.Spec.StorageClassName
		}
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, pvc.Namespace)
	}
	fields = append(fields,
		pvc.Name,
		string(phase),
		pvc.Spec.VolumeName,
		capacity,
		accessModes,
		class,
		toAge(pvc.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(pvc.ObjectMeta), fields

	return nil
}

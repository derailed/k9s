package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersistentVolumeClaim renders a K8s PersistentVolumeClaim to screen.
type PersistentVolumeClaim struct{}

// ColorerFunc colors a resource row.
func (p PersistentVolumeClaim) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, re)
		if re.Kind == EventAdd || re.Kind == EventUpdate {
			return c
		}
		if !Happy(ns, re.Row) {
			return ErrColor
		}

		return c
	}

}

// Header returns a header rbw.
func (PersistentVolumeClaim) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "VOLUME"},
		Header{Name: "CAPACITY"},
		Header{Name: "ACCESS MODES"},
		Header{Name: "STORAGECLASS"},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (p PersistentVolumeClaim) Render(o interface{}, ns string, r *Row) error {
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

	r.ID = client.MetaFQN(pvc.ObjectMeta)
	r.Fields = make(Fields, 0, len(p.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, pvc.Namespace)
	}
	r.Fields = append(r.Fields,
		pvc.Name,
		string(phase),
		pvc.Spec.VolumeName,
		capacity,
		accessModes,
		class,
		mapToStr(pvc.Labels),
		asStatus(p.diagnose(string(phase))),
		toAge(pvc.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (PersistentVolumeClaim) diagnose(r string) error {
	if r != "Bound" && r != "Available" {
		return fmt.Errorf("unexpected status %s", r)
	}
	return nil
}

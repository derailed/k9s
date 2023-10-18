package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersistentVolumeClaim renders a K8s PersistentVolumeClaim to screen.
type PersistentVolumeClaim struct {
	Base
}

// Header returns a header rbw.
func (PersistentVolumeClaim) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "VOLUME"},
		HeaderColumn{Name: "CAPACITY", Capacity: true},
		HeaderColumn{Name: "ACCESS MODES"},
		HeaderColumn{Name: "STORAGECLASS"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
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
	r.Fields = Fields{
		pvc.Namespace,
		pvc.Name,
		string(phase),
		pvc.Spec.VolumeName,
		capacity,
		accessModes,
		class,
		mapToStr(pvc.Labels),
		asStatus(p.diagnose(string(phase))),
		toAge(pvc.GetCreationTimestamp()),
	}

	return nil
}

func (PersistentVolumeClaim) diagnose(r string) error {
	if r != "Bound" && r != "Available" {
		return fmt.Errorf("unexpected status %s", r)
	}
	return nil
}

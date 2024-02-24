// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersistentVolumeClaim renders a K8s PersistentVolumeClaim to screen.
type PersistentVolumeClaim struct {
	Base
}

// Header returns a header rbw.
func (PersistentVolumeClaim) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "VOLUME"},
		model1.HeaderColumn{Name: "CAPACITY", Capacity: true},
		model1.HeaderColumn{Name: "ACCESS MODES"},
		model1.HeaderColumn{Name: "STORAGECLASS"},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (p PersistentVolumeClaim) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected PersistentVolumeClaim, but got %T", o)
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
	r.Fields = model1.Fields{
		pvc.Namespace,
		pvc.Name,
		string(phase),
		pvc.Spec.VolumeName,
		capacity,
		accessModes,
		class,
		mapToStr(pvc.Labels),
		AsStatus(p.diagnose(string(phase))),
		ToAge(pvc.GetCreationTimestamp()),
	}

	return nil
}

func (PersistentVolumeClaim) diagnose(r string) error {
	if r != "Bound" && r != "Available" {
		return fmt.Errorf("unexpected status %s", r)
	}
	return nil
}

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/util/storage"
)

// StorageClass renders a K8s StorageClass to screen.
type StorageClass struct {
	Base
}

// Header returns a header row.
func (StorageClass) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "PROVISIONER"},
		HeaderColumn{Name: "RECLAIMPOLICY"},
		HeaderColumn{Name: "VOLUMEBINDINGMODE"},
		HeaderColumn{Name: "ALLOWVOLUMEEXPANSION"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (s StorageClass) Render(o interface{}, ns string, r *Row) error {
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
		s.nameWithDefault(sc.ObjectMeta),
		sc.Provisioner,
		strPtrToStr((*string)(sc.ReclaimPolicy)),
		strPtrToStr((*string)(sc.VolumeBindingMode)),
		boolPtrToStr(sc.AllowVolumeExpansion),
		mapToStr(sc.Labels),
		"",
		toAge(sc.GetCreationTimestamp()),
	}

	return nil
}

func (StorageClass) nameWithDefault(meta metav1.ObjectMeta) string {
	if storage.IsDefaultAnnotationText(meta) == "Yes" {
		return meta.Name + " (default)"
	}
	return meta.Name
}

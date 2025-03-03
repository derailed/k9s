// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
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
func (s StorageClass) Header(_ string) model1.Header {
	return s.doHeader(s.defaultHeader())
}

func (StorageClass) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "PROVISIONER"},
		model1.HeaderColumn{Name: "RECLAIMPOLICY"},
		model1.HeaderColumn{Name: "VOLUMEBINDINGMODE"},
		model1.HeaderColumn{Name: "ALLOWVOLUMEEXPANSION"},
		model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (s StorageClass) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected StorageClass, but got %T", o)
	}

	if err := s.defaultRow(raw, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}

	cols, err := s.specs.realize(raw, s.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (s StorageClass) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var sc storagev1.StorageClass
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sc)
	if err != nil {
		return err
	}

	r.ID = client.FQN(client.ClusterScope, sc.ObjectMeta.Name)
	r.Fields = model1.Fields{
		s.nameWithDefault(sc.ObjectMeta),
		sc.Provisioner,
		strPtrToStr((*string)(sc.ReclaimPolicy)),
		strPtrToStr((*string)(sc.VolumeBindingMode)),
		boolPtrToStr(sc.AllowVolumeExpansion),
		mapToStr(sc.Labels),
		"",
		ToAge(sc.GetCreationTimestamp()),
	}

	return nil
}

func (StorageClass) nameWithDefault(meta metav1.ObjectMeta) string {
	if storage.IsDefaultAnnotationText(meta) == "Yes" {
		return meta.Name + " (default)"
	}
	return meta.Name
}

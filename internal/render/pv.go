// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const terminatingPhase = "Terminating"

// PersistentVolume renders a K8s PersistentVolume to screen.
type PersistentVolume struct {
	Base
}

// ColorerFunc colors a resource row.
func (p PersistentVolume) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("STATUS", true)
		if ok {
			return c
		}
		switch strings.TrimSpace(re.Row.Fields[idx]) {
		case string(v1.VolumeBound):
			return model1.StdColor
		case string(v1.VolumeAvailable):
			return tcell.ColorGreen
		case string(v1.VolumePending):
			return model1.PendingColor
		case terminatingPhase:
			return model1.CompletedColor
		}

		return c
	}
}

// Header returns a header rbw.
func (PersistentVolume) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "CAPACITY", Capacity: true},
		model1.HeaderColumn{Name: "ACCESS MODES"},
		model1.HeaderColumn{Name: "RECLAIM POLICY"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "CLAIM"},
		model1.HeaderColumn{Name: "STORAGECLASS"},
		model1.HeaderColumn{Name: "REASON"},
		model1.HeaderColumn{Name: "VOLUMEMODE", Wide: true},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (p PersistentVolume) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected PersistentVolume, but got %T", o)
	}
	var pv v1.PersistentVolume
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &pv)
	if err != nil {
		return err
	}

	phase := pv.Status.Phase
	if pv.ObjectMeta.DeletionTimestamp != nil {
		phase = terminatingPhase
	}
	var claim string
	if pv.Spec.ClaimRef != nil {
		claim = path.Join(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}
	class, found := pv.Annotations[v1.BetaStorageClassAnnotation]
	if !found {
		class = pv.Spec.StorageClassName
	}

	size := pv.Spec.Capacity[v1.ResourceStorage]

	r.ID = client.MetaFQN(pv.ObjectMeta)
	r.Fields = model1.Fields{
		pv.Name,
		size.String(),
		accessMode(pv.Spec.AccessModes),
		string(pv.Spec.PersistentVolumeReclaimPolicy),
		string(phase),
		claim,
		class,
		pv.Status.Reason,
		p.volumeMode(pv.Spec.VolumeMode),
		mapToStr(pv.Labels),
		AsStatus(p.diagnose(phase)),
		ToAge(pv.GetCreationTimestamp()),
	}

	return nil
}

func (PersistentVolume) diagnose(phase v1.PersistentVolumePhase) error {
	if phase == v1.VolumeFailed {
		return fmt.Errorf("failed to delete or recycle")
	}

	return nil
}

func (PersistentVolume) volumeMode(m *v1.PersistentVolumeMode) string {
	if m == nil {
		return MissingValue
	}

	return string(*m)
}

// ----------------------------------------------------------------------------
// Helpers...

func accessMode(aa []v1.PersistentVolumeAccessMode) string {
	dd := accessDedup(aa)
	s := make([]string, 0, len(dd))
	for _, am := range dd {
		switch am {
		case v1.ReadWriteOnce:
			s = append(s, "RWO")
		case v1.ReadOnlyMany:
			s = append(s, "ROX")
		case v1.ReadWriteMany:
			s = append(s, "RWX")
		case v1.ReadWriteOncePod:
			s = append(s, "RWOP")
		}
	}

	return strings.Join(s, ",")
}

func accessContains(cc []v1.PersistentVolumeAccessMode, a v1.PersistentVolumeAccessMode) bool {
	for _, c := range cc {
		if c == a {
			return true
		}
	}

	return false
}

func accessDedup(cc []v1.PersistentVolumeAccessMode) []v1.PersistentVolumeAccessMode {
	set := []v1.PersistentVolumeAccessMode{}
	for _, c := range cc {
		if !accessContains(set, c) {
			set = append(set, c)
		}
	}

	return set
}

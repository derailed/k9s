// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/client"
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
func (p PersistentVolume) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		if !Happy(ns, h, re.Row) {
			return ErrColor
		}

		statusCol := h.IndexOf("STATUS", true)
		if statusCol == -1 {
			return DefaultColorer(ns, h, re)
		}
		switch strings.TrimSpace(re.Row.Fields[statusCol]) {
		case string(v1.VolumeBound):
			return StdColor
		case string(v1.VolumeAvailable):
			return tcell.ColorGreen
		case string(v1.VolumePending):
			return PendingColor
		case terminatingPhase:
			return CompletedColor
		}

		return DefaultColorer(ns, h, re)
	}
}

// Header returns a header rbw.
func (PersistentVolume) Header(string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "CAPACITY", Capacity: true},
		HeaderColumn{Name: "ACCESS MODES"},
		HeaderColumn{Name: "RECLAIM POLICY"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "CLAIM"},
		HeaderColumn{Name: "STORAGECLASS"},
		HeaderColumn{Name: "REASON"},
		HeaderColumn{Name: "VOLUMEMODE", Wide: true},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (p PersistentVolume) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected PersistentVolume, but got %T", o)
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
	r.Fields = Fields{
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
		asStatus(p.diagnose(phase)),
		toAge(pv.GetCreationTimestamp()),
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

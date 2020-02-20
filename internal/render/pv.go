package render

import (
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PersistentVolume renders a K8s PersistentVolume to screen.
type PersistentVolume struct{}

// ColorerFunc colors a resource row.
func (p PersistentVolume) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, re)
		if re.Kind == EventAdd || re.Kind == EventUpdate {
			return c
		}

		if !Happy(ns, re.Row) {
			return ErrColor
		}

		switch strings.TrimSpace(re.Row.Fields[4]) {
		case "Bound":
			c = StdColor
		case "Available":
			c = tcell.ColorYellow
		}

		return c
	}
}

// Header returns a header rbw.
func (PersistentVolume) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "CAPACITY"},
		Header{Name: "ACCESS MODES"},
		Header{Name: "RECLAIM POLICY"},
		Header{Name: "STATUS"},
		Header{Name: "CLAIM"},
		Header{Name: "STORAGECLASS"},
		Header{Name: "REASON"},
		Header{Name: "VOLUMEMODE", Wide: true},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
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
		phase = "Terminating"
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
		asStatus(p.diagnose(string(phase))),
		toAge(pv.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (PersistentVolume) diagnose(r string) error {
	if r != "Bound" && r != "Available" {
		return fmt.Errorf("unexpected status %s", r)
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
	for i := 0; i < len(aa); i++ {
		switch {
		case accessContains(dd, v1.ReadWriteOnce):
			s = append(s, "RWO")
		case accessContains(dd, v1.ReadOnlyMany):
			s = append(s, "ROX")
		case accessContains(dd, v1.ReadWriteMany):
			s = append(s, "RWX")
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

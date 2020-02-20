package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodDisruptionBudget renders a K8s PodDisruptionBudget to screen.
type PodDisruptionBudget struct{}

// ColorerFunc colors a resource row.
func (p PodDisruptionBudget) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, re)
		if re.Kind == EventAdd || re.Kind == EventUpdate {
			return c
		}

		if !Happy(ns, re.Row) {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (PodDisruptionBudget) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "MIN AVAILABLE", Align: tview.AlignRight},
		Header{Name: "MAX_ UNAVAILABLE", Align: tview.AlignRight},
		Header{Name: "ALLOWED DISRUPTIONS", Align: tview.AlignRight},
		Header{Name: "CURRENT", Align: tview.AlignRight},
		Header{Name: "DESIRED", Align: tview.AlignRight},
		Header{Name: "EXPECTED", Align: tview.AlignRight},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (p PodDisruptionBudget) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected PodDisruptionBudget, but got %T", o)
	}
	var pdb v1beta1.PodDisruptionBudget
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &pdb)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(pdb.ObjectMeta)
	r.Fields = make(Fields, 0, len(p.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, pdb.Namespace)
	}
	r.Fields = append(r.Fields,
		pdb.Name,
		numbToStr(pdb.Spec.MinAvailable),
		numbToStr(pdb.Spec.MaxUnavailable),
		strconv.Itoa(int(pdb.Status.PodDisruptionsAllowed)),
		strconv.Itoa(int(pdb.Status.CurrentHealthy)),
		strconv.Itoa(int(pdb.Status.DesiredHealthy)),
		strconv.Itoa(int(pdb.Status.ExpectedPods)),
		mapToStr(pdb.Labels),
		asStatus(p.diagnose(pdb.Spec.MinAvailable, pdb.Status.CurrentHealthy)),
		toAge(pdb.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (PodDisruptionBudget) diagnose(min *intstr.IntOrString, healthy int32) error {
	if min == nil {
		return nil
	}
	if min.IntVal > healthy {
		return fmt.Errorf("expected %d but got %d", min.IntVal, healthy)
	}
	return nil
}

// Helpers...

func numbToStr(n *intstr.IntOrString) string {
	if n == nil {
		return NAValue
	}
	return strconv.Itoa(int(n.IntVal))
}

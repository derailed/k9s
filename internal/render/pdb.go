package render

import (
	"fmt"
	"strconv"
	"strings"

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
func (PodDisruptionBudget) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}

		markCol := 5
		if client.IsNamespaced(ns) {
			markCol = 4
		}
		if strings.TrimSpace(r.Row.Fields[markCol]) != strings.TrimSpace(r.Row.Fields[markCol+1]) {
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

	r.ID = MetaFQN(pdb.ObjectMeta)
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
		toAge(pdb.ObjectMeta.CreationTimestamp),
	)

	return nil
}

// Helpers...

func numbToStr(n *intstr.IntOrString) string {
	if n == nil {
		return NAValue
	}
	return strconv.Itoa(int(n.IntVal))
}

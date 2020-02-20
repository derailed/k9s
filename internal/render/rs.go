package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ReplicaSet renders a K8s ReplicaSet to screen.
type ReplicaSet struct{}

// ColorerFunc colors a resource row.
func (r ReplicaSet) ColorerFunc() ColorerFunc {
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
func (ReplicaSet) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DESIRED", Align: tview.AlignRight},
		Header{Name: "CURRENT", Align: tview.AlignRight},
		Header{Name: "READY", Align: tview.AlignRight},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s ReplicaSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ReplicaSet, but got %T", o)
	}
	var rs appsv1.ReplicaSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rs)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(rs.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, rs.Namespace)
	}
	r.Fields = append(r.Fields,
		rs.Name,
		strconv.Itoa(int(*rs.Spec.Replicas)),
		strconv.Itoa(int(rs.Status.Replicas)),
		strconv.Itoa(int(rs.Status.ReadyReplicas)),
		mapToStr(rs.Labels),
		asStatus(s.diagnose(rs)),
		toAge(rs.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (s ReplicaSet) diagnose(rs appsv1.ReplicaSet) error {
	if rs.Status.Replicas != rs.Status.ReadyReplicas {
		if rs.Status.Replicas == 0 {
			return fmt.Errorf("did not phase down correctly expecting 0 replicas but got %d", rs.Status.ReadyReplicas)
		}
		return fmt.Errorf("mismatch desired(%d) vs ready(%d)", rs.Status.Replicas, rs.Status.ReadyReplicas)
	}

	return nil
}

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ReplicaSet renders a K8s ReplicaSet to screen.
type ReplicaSet struct{}

// ColorerFunc colors a resource row.
func (ReplicaSet) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (ReplicaSet) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DESIRED", Align: tview.AlignRight},
		Header{Name: "CURRENT", Align: tview.AlignRight},
		Header{Name: "READY", Align: tview.AlignRight},
		Header{Name: "AGE"},
	)
}

// Render renders a K8s resource to screen.
func (ReplicaSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ReplicaSet, but got %T", o)
	}
	var rs appsv1.ReplicaSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rs)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, rs.Namespace)
	}
	fields = append(fields,
		rs.Name,
		strconv.Itoa(int(*rs.Spec.Replicas)),
		strconv.Itoa(int(rs.Status.Replicas)),
		strconv.Itoa(int(rs.Status.ReadyReplicas)),
		toAge(rs.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(rs.ObjectMeta), fields

	return nil
}

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Deployment renders a K8s Deployment to screen.
type Deployment struct{}

func isAllNamespace(ns string) bool {
	return ns == ""
}

// ColorerFunc colors a resource row.
func (Deployment) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Deployment) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DESIRED", Align: tview.AlignRight},
		Header{Name: "CURRENT", Align: tview.AlignRight},
		Header{Name: "UP-TO-DATE", Align: tview.AlignRight},
		Header{Name: "AVAILABLE", Align: tview.AlignRight},
		Header{Name: "AGE"},
	)
}

// Render renders a K8s resource to screen.
func (Deployment) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Deployment, but got %T", o)
	}
	var dp appsv1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &dp)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, dp.Namespace)
	}
	fields = append(fields,
		dp.Name,
		strconv.Itoa(int(*dp.Spec.Replicas)),
		strconv.Itoa(int(dp.Status.Replicas)),
		strconv.Itoa(int(dp.Status.UpdatedReplicas)),
		strconv.Itoa(int(dp.Status.AvailableReplicas)),
		toAge(dp.ObjectMeta.CreationTimestamp),
	)

	r.ID, r.Fields = MetaFQN(dp.ObjectMeta), fields

	return nil
}

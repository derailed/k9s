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

// Deployment renders a K8s Deployment to screen.
type Deployment struct{}

// ColorerFunc colors a resource row.
func (d Deployment) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}
		if !Happy(ns, r.Row) {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (Deployment) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "READY"},
		Header{Name: "UP-TO-DATE", Align: tview.AlignRight},
		Header{Name: "AVAILABLE", Align: tview.AlignRight},
		Header{Name: "READY", Align: tview.AlignRight},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (d Deployment) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Deployment, but got %T", o)
	}

	var dp appsv1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &dp)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(dp.ObjectMeta)
	r.Fields = make(Fields, 0, len(d.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, dp.Namespace)
	}
	r.Fields = append(r.Fields,
		dp.Name,
		strconv.Itoa(int(dp.Status.AvailableReplicas))+"/"+strconv.Itoa(int(dp.Status.Replicas)),
		strconv.Itoa(int(dp.Status.UpdatedReplicas)),
		strconv.Itoa(int(dp.Status.AvailableReplicas)),
		strconv.Itoa(int(dp.Status.ReadyReplicas)),
		mapToStr(dp.Labels),
		asStatus(d.diagnose(dp.Status.Replicas, dp.Status.AvailableReplicas)),
		toAge(dp.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (Deployment) diagnose(d, r int32) error {
	if d != r {
		return fmt.Errorf("desiring %d replicas got %d available", d, r)
	}
	return nil
}

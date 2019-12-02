package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}

		markCol := 2
		if ns != AllNamespaces {
			markCol = 1
		}
		tokens := strings.Split(r.Row.Fields[markCol], "/")
		if tokens[0] != tokens[1] {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (Deployment) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "READY"},
		Header{Name: "UP-TO-DATE", Align: tview.AlignRight},
		Header{Name: "AVAILABLE", Align: tview.AlignRight},
		Header{Name: "SELECTOR"},
		Header{Name: "AGE", Decorator: ageDecorator},
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

	r.ID = MetaFQN(dp.ObjectMeta)
	r.Fields = make(Fields, 0, len(d.Header(ns)))
	if isAllNamespace(ns) {
		r.Fields = append(r.Fields, dp.Namespace)
	}
	r.Fields = append(r.Fields,
		dp.Name,
		strconv.Itoa(int(dp.Status.AvailableReplicas))+"/"+strconv.Itoa(int(*dp.Spec.Replicas)),
		strconv.Itoa(int(dp.Status.UpdatedReplicas)),
		strconv.Itoa(int(dp.Status.AvailableReplicas)),
		asSelector(dp.Spec.Selector),
		toAge(dp.ObjectMeta.CreationTimestamp),
	)

	return nil
}

//Helpers...

func asSelector(s *metav1.LabelSelector) string {
	sel, err := metav1.LabelSelectorAsSelector(s)
	if err != nil {
		log.Error().Err(err).Msg("Selector conversion failed")
		return NAValue
	}

	return sel.String()
}

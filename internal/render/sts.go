package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulSet renders a K8s StatefulSet to screen.
type StatefulSet struct{}

// ColorerFunc colors a resource row.
func (StatefulSet) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}

		readyCol := 2
		if !client.IsAllNamespaces(ns) {
			readyCol--
		}
		tokens := strings.Split(strings.TrimSpace(r.Row.Fields[readyCol]), "/")
		curr, des := tokens[0], tokens[1]
		if curr != des {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (StatefulSet) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "READY"},
		Header{Name: "SELECTOR"},
		Header{Name: "SERVICE"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s StatefulSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected StatefulSet, but got %T", o)
	}
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sts.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, sts.Namespace)
	}
	r.Fields = append(r.Fields,
		sts.Name,
		strconv.Itoa(int(sts.Status.Replicas))+"/"+strconv.Itoa(int(*sts.Spec.Replicas)),
		asSelector(sts.Spec.Selector),
		na(sts.Spec.ServiceName),
		toAge(sts.ObjectMeta.CreationTimestamp),
	)

	return nil
}

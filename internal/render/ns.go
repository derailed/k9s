package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace renders a K8s Namespace to screen.
type Namespace struct{}

// ColorerFunc colors a resource row.
func (Namespace) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)

		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}
		switch strings.TrimSpace(r.Row.Fields[1]) {
		case "Inactive", Terminating:
			c = ErrColor
		}
		if strings.Contains(strings.TrimSpace(r.Row.Fields[0]), "*") {
			c = HighlightColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Namespace) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "AGE", Decorator: ageDecorator},
	}
}

// Render renders a K8s resource to screen.
func (Namespace) Render(o interface{}, _ string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Namespace, but got %T", o)
	}
	var ns v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ns)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	fields = append(fields,
		ns.Name,
		string(ns.Status.Phase),
		toAge(ns.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(ns.ObjectMeta), fields

	return nil
}

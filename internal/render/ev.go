package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Event renders a K8s Event to screen.
type Event struct{}

// ColorerFunc colors a resource row.
func (Event) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)

		markCol := 3
		if ns != AllNamespaces {
			markCol = 2
		}
		switch strings.TrimSpace(r.Row.Fields[markCol]) {
		case "Failed":
			c = ErrColor
		case "Killing":
			c = KillColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Event) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "REASON"},
		Header{Name: "SOURCE"},
		Header{Name: "COUNT", Align: tview.AlignRight},
		Header{Name: "MESSAGE"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (e Event) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Event, but got %T", o)
	}
	var ev v1.Event
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ev)
	if err != nil {
		return err
	}

	r.ID = MetaFQN(ev.ObjectMeta)
	r.Fields = make(Fields, 0, len(e.Header(ns)))
	if isAllNamespace(ns) {
		r.Fields = append(r.Fields, ev.Namespace)
	}
	r.Fields = append(r.Fields,
		ev.Name,
		ev.Reason,
		ev.Source.Component,
		strconv.Itoa(int(ev.Count)),
		Truncate(ev.Message, 80),
		toAge(ev.LastTimestamp))

	return nil
}

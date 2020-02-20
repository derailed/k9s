package render

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Event renders a K8s Event to screen.
type Event struct{}

// ColorerFunc colors a resource row.
func (e Event) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)

		if !Happy(ns, r.Row) {
			return ErrColor
		}

		markCol := 3
		if !client.IsAllNamespaces(ns) {
			markCol = 2
		}
		if strings.TrimSpace(r.Row.Fields[markCol]) == "Killing" {
			return KillColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Event) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "TYPE"},
		Header{Name: "REASON"},
		Header{Name: "SOURCE"},
		Header{Name: "COUNT", Align: tview.AlignRight},
		Header{Name: "MESSAGE", Wide: true},
		Header{Name: "VALID", Wide: true},
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

	r.ID = client.MetaFQN(ev.ObjectMeta)
	r.Fields = make(Fields, 0, len(e.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, ev.Namespace)
	}
	r.Fields = append(r.Fields,
		asRef(ev.InvolvedObject),
		ev.Type,
		ev.Reason,
		ev.Source.Component,
		strconv.Itoa(int(ev.Count)),
		ev.Message,
		asStatus(e.diagnose(ev.Type)),
		toAge(ev.LastTimestamp))

	return nil
}

// Happy returns true if resoure is happy, false otherwise
func (Event) diagnose(kind string) error {
	if kind != "Normal" {
		return errors.New("failed event")
	}
	return nil
}

// Helpers...

func asRef(r v1.ObjectReference) string {
	return strings.ToLower(r.Kind) + ":" + r.Name
}

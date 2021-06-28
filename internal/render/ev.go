package render

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Event renders a K8s Event to screen.
type Event struct{}

// ColorerFunc colors a resource row.
func (e Event) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		if !Happy(ns, h, re.Row) {
			return ErrColor
		}
		reasonCol := h.IndexOf("REASON", true)
		if reasonCol == -1 {
			return DefaultColorer(ns, h, re)
		}
		if strings.TrimSpace(re.Row.Fields[reasonCol]) == "Killing" {
			return KillColor
		}

		return DefaultColorer(ns, h, re)
	}
}

// Header returns a header rbw.
func (Event) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "TYPE"},
		HeaderColumn{Name: "REASON"},
		HeaderColumn{Name: "SOURCE"},
		HeaderColumn{Name: "COUNT", Align: tview.AlignRight},
		HeaderColumn{Name: "MESSAGE", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
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
	r.Fields = Fields{
		ev.Namespace,
		asRef(ev.InvolvedObject),
		ev.Type,
		ev.Reason,
		ev.Source.Component,
		strconv.Itoa(int(ev.Count)),
		ev.Message,
		asStatus(e.diagnose(ev.Type)),
		toAge(ev.LastTimestamp),
	}

	return nil
}

// Happy returns true if resource is happy, false otherwise.
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

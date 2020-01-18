package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// DaemonSet renders a K8s DaemonSet to screen.
type DaemonSet struct{}

// ColorerFunc colors a resource row.
func (DaemonSet) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd || r.Kind == EventUpdate {
			return c
		}

		desiredCol := 2
		if !client.IsAllNamespaces(ns) {
			desiredCol = 1
		}
		if strings.TrimSpace(r.Row.Fields[desiredCol]) != strings.TrimSpace(r.Row.Fields[desiredCol+2]) {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (DaemonSet) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DESIRED", Align: tview.AlignRight},
		Header{Name: "CURRENT", Align: tview.AlignRight},
		Header{Name: "READY", Align: tview.AlignRight},
		Header{Name: "UP-TO-DATE", Align: tview.AlignRight},
		Header{Name: "AVAILABLE", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (d DaemonSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected DaemonSet, but got %T", o)
	}
	var ds appsv1.DaemonSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ds)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(ds.ObjectMeta)
	r.Fields = make(Fields, 0, len(d.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, ds.Namespace)
	}
	r.Fields = append(r.Fields,
		ds.Name,
		strconv.Itoa(int(ds.Status.DesiredNumberScheduled)),
		strconv.Itoa(int(ds.Status.CurrentNumberScheduled)),
		strconv.Itoa(int(ds.Status.NumberReady)),
		strconv.Itoa(int(ds.Status.UpdatedNumberScheduled)),
		strconv.Itoa(int(ds.Status.NumberAvailable)),
		toAge(ds.ObjectMeta.CreationTimestamp),
	)

	return nil
}

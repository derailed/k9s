package render

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace renders a K8s Namespace to screen.
type Namespace struct{}

// ColorerFunc colors a resource row.
func (n Namespace) ColorerFunc() ColorerFunc {
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)
		if r.Kind == EventAdd {
			return c
		}

		if r.Kind == EventUpdate {
			c = StdColor
		}
		if !Happy(ns, r.Row) {
			return ErrColor
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
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (n Namespace) Render(o interface{}, _ string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Namespace, but got %T", o)
	}
	var ns v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ns)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(ns.ObjectMeta)
	r.Fields = Fields{
		ns.Name,
		string(ns.Status.Phase),
		mapToStr(ns.Labels),
		asStatus(n.diagnose(ns.Status.Phase)),
		toAge(ns.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (Namespace) diagnose(phase v1.NamespacePhase) error {
	if phase != v1.NamespaceActive && phase != v1.NamespaceTerminating {
		return errors.New("namespace not ready")
	}
	return nil
}

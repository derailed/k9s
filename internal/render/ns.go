package render

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace renders a K8s Namespace to screen.
type Namespace struct {
	Base
}

// ColorerFunc colors a resource row.
func (n Namespace) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, h, re)

		if re.Kind == EventUpdate {
			c = StdColor
		}
		if strings.Contains(strings.TrimSpace(re.Row.Fields[0]), "*") {
			c = HighlightColor
		}

		if !Happy(ns, h, re.Row) {
			c = ErrColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Namespace) Header(string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
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
		toAge(ns.GetCreationTimestamp()),
	}

	return nil
}

func (Namespace) diagnose(phase v1.NamespacePhase) error {
	if phase != v1.NamespaceActive && phase != v1.NamespaceTerminating {
		return errors.New("namespace not ready")
	}
	return nil
}

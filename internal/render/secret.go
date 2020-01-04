package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Secret renders a K8s Secret to screen.
type Secret struct{}

// ColorerFunc colors a resource row.
func (Secret) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Secret) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "TYPE"},
		Header{Name: "DATA", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s Secret) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Secret, but got %T", o)
	}
	var sec v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sec)
	if err != nil {
		return err
	}

	r.ID = MetaFQN(sec.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, sec.Namespace)
	}
	r.Fields = append(r.Fields,
		sec.Name,
		string(sec.Type),
		strconv.Itoa(len(sec.Data)),
		toAge(sec.ObjectMeta.CreationTimestamp),
	)

	return nil
}

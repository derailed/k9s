package render

import (
	"fmt"
	"strconv"

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
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "TYPE"},
		Header{Name: "DATA", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: ageDecorator},
	)
}

// Render renders a K8s resource to screen.
func (Secret) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Secret, but got %T", o)
	}
	var s v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &s)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, s.Namespace)
	}
	fields = append(fields,
		s.Name,
		string(s.Type),
		strconv.Itoa(len(s.Data)),
		toAge(s.ObjectMeta.CreationTimestamp),
	)

	r.ID, r.Fields = MetaFQN(s.ObjectMeta), fields

	return nil
}

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConfigMap renders a K8s ConfigMap to screen.
type ConfigMap struct{}

// ColorerFunc colors a resource row.
func (ConfigMap) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (ConfigMap) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DATA", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: ageDecorator},
	)
}

// Render renders a K8s resource to screen.
func (ConfigMap) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ConfigMap, but got %T", o)
	}
	var cm v1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cm)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, cm.Namespace)
	}
	fields = append(fields,
		cm.Name,
		strconv.Itoa(len(cm.Data)),
		toAge(cm.ObjectMeta.CreationTimestamp),
	)

	r.ID, r.Fields = MetaFQN(cm.ObjectMeta), fields

	return nil
}

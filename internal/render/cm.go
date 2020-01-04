package render

import (
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "DATA", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
// BOZO!! 44allocs down to 5allocs avoiding marshal??
func (c ConfigMap) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ConfigMap, but got %T", o)
	}

	meta, ok := raw.Object["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("No meta")
	}

	n, nss := extractMetaField(meta, "name"), extractMetaField(meta, "namespace")
	r.ID = FQN(nss, n)
	r.Fields = make(Fields, 0, len(c.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, nss)
	}

	var size int
	data, ok := raw.Object["data"]
	if ok {
		d, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expecting map but got %T", raw.Object["data"])
		}
		size = len(d)
	}
	t, err := extractMetaTime(meta)
	if err != nil {
		return err
	}
	r.Fields = append(r.Fields,
		n,
		strconv.Itoa(size),
		toAge(t),
	)

	// var cm v1.ConfigMap
	// err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cm)
	// if err != nil {
	// 	return err
	// }

	// r.ID = MetaFQN(cm.ObjectMeta)
	// r.Fields = make(Fields, 0, len(c.Header(ns)))
	// if client.IsAllNamespaces(ns) {
	// 	r.Fields = append(r.Fields, cm.Namespace)
	// }
	// r.Fields = append(r.Fields,
	// 	cm.Name,
	// 	strconv.Itoa(len(cm.Data)),
	// 	toAge(cm.ObjectMeta.CreationTimestamp),
	// )

	return nil
}

func extractMetaTime(m map[string]interface{}) (metav1.Time, error) {
	f, ok := m["creationTimestamp"]
	if !ok {
		return metav1.Time{}, fmt.Errorf("failed to extract time from meta")
	}

	t, ok := f.(string)
	if !ok {
		return metav1.Time{}, fmt.Errorf("failed to extract time from field")
	}

	ti, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return metav1.Time{}, err
	}
	return metav1.Time{Time: ti}, nil
}

package render

import (
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	ofaas "github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	fnStatusReady    = "Ready"
	fnStatusNotReady = "Not Ready"
)

// OpenFaas renders an openfaas function to screen.
type OpenFaas struct{}

// ColorerFunc colors a resource row.
func (OpenFaas) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorPaleTurquoise
	}
}

// Header returns a header row.
func (OpenFaas) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "IMAGE"},
		Header{Name: "LABELS"},
		Header{Name: "INVOCATIONS", Align: tview.AlignRight},
		Header{Name: "REPLICAS", Align: tview.AlignRight},
		Header{Name: "AVAILABLE", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a chart to screen.
func (f OpenFaas) Render(o interface{}, ns string, r *Row) error {
	fn, ok := o.(OpenFaasRes)
	if !ok {
		return fmt.Errorf("expected OpenFaasRes, but got %T", o)
	}

	var labels string
	if fn.Function.Labels != nil {
		labels = mapToStr(*fn.Function.Labels)
	}
	var status = fnStatusReady
	if fn.Function.AvailableReplicas == 0 {
		status = fnStatusNotReady
	}

	r.ID = client.FQN(fn.Function.Namespace, fn.Function.Name)
	r.Fields = make(Fields, 0, len(f.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, fn.Function.Namespace)
	}
	r.Fields = append(r.Fields,
		fn.Function.Name,
		status,
		fn.Function.Image,
		labels,
		strconv.Itoa(int(fn.Function.InvocationCount)),
		strconv.Itoa(int(fn.Function.Replicas)),
		strconv.Itoa(int(fn.Function.AvailableReplicas)),
		toAge(metav1.Time{Time: time.Now()}),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// OpenFaasRes represents an openfaas function resource.
type OpenFaasRes struct {
	Function ofaas.FunctionStatus `json:"function"`
}

// GetObjectKind returns a schema object.
func (OpenFaasRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (h OpenFaasRes) DeepCopyObject() runtime.Object {
	return h
}

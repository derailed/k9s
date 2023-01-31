package render

// BOZO!! revamp with latest...

// import (
// 	"errors"
// 	"fmt"
// 	"strconv"
// 	"time"

// 	"github.com/derailed/k9s/internal/client"
// 	"github.com/derailed/tview"
// 	"github.com/derailed/tcell/v2"

// 	ofaas "github.com/openfaas/faas-provider/types"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// )

// const (
// 	fnStatusReady    = "Ready"
// 	fnStatusNotReady = "Not Ready"
// )

// // OpenFaas renders an openfaas function to screen.
// type OpenFaas struct{}

// // ColorerFunc colors a resource row.
// func (o OpenFaas) ColorerFunc() ColorerFunc {
// 	return func(ns string, h Header, re RowEvent) tcell.Color {
// 		if !Happy(ns, h, re.Row) {
// 			return ErrColor
// 		}

// 		return tcell.ColorPaleTurquoise
// 	}
// }

// // Header returns a header row.
// func (OpenFaas) Header(ns string) Header {
// 	return Header{
// 		HeaderColumn{Name: "NAMESPACE"},
// 		HeaderColumn{Name: "NAME"},
// 		HeaderColumn{Name: "STATUS"},
// 		HeaderColumn{Name: "IMAGE"},
// 		HeaderColumn{Name: "LABELS"},
// 		HeaderColumn{Name: "INVOCATIONS", Align: tview.AlignRight},
// 		HeaderColumn{Name: "REPLICAS", Align: tview.AlignRight},
// 		HeaderColumn{Name: "AVAILABLE", Align: tview.AlignRight},
// 		HeaderColumn{Name: "VALID", Wide: true},
// 		HeaderColumn{Name: "AGE", Time: true},
// 	}
// }

// // Render renders a chart to screen.
// func (o OpenFaas) Render(i interface{}, ns string, r *Row) error {
// 	fn, ok := i.(OpenFaasRes)
// 	if !ok {
// 		return fmt.Errorf("expected OpenFaasRes, but got %T", o)
// 	}

// 	var labels string
// 	if fn.Function.Labels != nil {
// 		labels = mapToStr(*fn.Function.Labels)
// 	}
// 	status := fnStatusReady
// 	if fn.Function.AvailableReplicas == 0 {
// 		status = fnStatusNotReady
// 	}

// 	r.ID = client.FQN(fn.Function.Namespace, fn.Function.Name)
// 	r.Fields = Fields{
// 		fn.Function.Namespace,
// 		fn.Function.Name,
// 		status,
// 		fn.Function.Image,
// 		labels,
// 		strconv.Itoa(int(fn.Function.InvocationCount)),
// 		strconv.Itoa(int(fn.Function.Replicas)),
// 		strconv.Itoa(int(fn.Function.AvailableReplicas)),
// 		asStatus(o.diagnose(status)),
// 		toAge(metav1.Time{Time: time.Now()}),
// 	}

// 	return nil
// }

// func (OpenFaas) diagnose(status string) error {
// 	if status != "Ready" {
// 		return errors.New("function not ready")
// 	}

// 	return nil
// }

// // ----------------------------------------------------------------------------
// // Helpers...

// // OpenFaasRes represents an openfaas function resource.
// type OpenFaasRes struct {
// 	Function ofaas.FunctionStatus `json:"function"`
// }

// // GetObjectKind returns a schema object.
// func (OpenFaasRes) GetObjectKind() schema.ObjectKind {
// 	return nil
// }

// // DeepCopyObject returns a container copy.
// func (h OpenFaasRes) DeepCopyObject() runtime.Object {
// 	return h
// }

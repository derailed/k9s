package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// BOZO!!
// import (
// 	"errors"
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/derailed/k9s/internal/client"
// 	"github.com/derailed/tview"
// 	"github.com/gdamore/tcell/v2"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/util/duration"
// 	api "k8s.io/kubernetes/pkg/apis/core"
// )

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

// // Header returns a header rbw.
// func (Event) Header(ns string) Header {
// 	return Header{
// 		HeaderColumn{Name: "NAMESPACE"},
// 		HeaderColumn{Name: "LAST SEEN"},
// 		HeaderColumn{Name: "TYPE"},
// 		HeaderColumn{Name: "REASON"},
// 		HeaderColumn{Name: "OBJECT"},
// 		HeaderColumn{Name: "SUBOBJECT"},
// 		HeaderColumn{Name: "SOURCE"},
// 		HeaderColumn{Name: "MESSAGE", Wide: true},
// 		HeaderColumn{Name: "FIRST SEEN", Wide: true},
// 		HeaderColumn{Name: "COUNT", Align: tview.AlignRight},
// 		HeaderColumn{Name: "NAME"},
// 		HeaderColumn{Name: "VALID", Wide: true},
// 	}
// }

// // Render renders a K8s resource to screen.
// func (e Event) Render(o interface{}, ns string, r *Row) error {
// 	raw, ok := o.(*unstructured.Unstructured)
// 	if !ok {
// 		return fmt.Errorf("Expected Event, but got %T", o)
// 	}
// 	var ev api.Event
// 	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ev)
// 	if err != nil {
// 		return err
// 	}

// 	firstTimestamp := translateTimestampSince(ev.FirstTimestamp)
// 	if ev.FirstTimestamp.IsZero() {
// 		firstTimestamp = translateMicroTimestampSince(ev.EventTime)
// 	}

// 	lastTimestamp := translateTimestampSince(ev.LastTimestamp)
// 	if ev.LastTimestamp.IsZero() {
// 		lastTimestamp = firstTimestamp
// 	}
// 	count := ev.Count
// 	if ev.Series != nil {
// 		lastTimestamp = translateMicroTimestampSince(ev.Series.LastObservedTime)
// 		count = ev.Series.Count
// 	} else if count == 0 {
// 		// Singleton events don't have a count set in the new API.
// 		count = 1
// 	}

// 	var target string
// 	if len(ev.InvolvedObject.Name) > 0 {
// 		target = fmt.Sprintf("%s/%s", strings.ToLower(ev.InvolvedObject.Kind), ev.InvolvedObject.Name)
// 	} else {
// 		target = strings.ToLower(ev.InvolvedObject.Kind)
// 	}

// 	r.ID = client.MetaFQN(ev.ObjectMeta)
// 	r.Fields = Fields{
// 		ev.Namespace,
// 		lastTimestamp,
// 		ev.Type,
// 		ev.Reason,
// 		target,
// 		ev.InvolvedObject.FieldPath,
// 		fmtEventSource(ev.Source, ev.ReportingController, ev.ReportingInstance),
// 		strings.TrimSpace(ev.Message),
// 		firstTimestamp,
// 		strconv.Itoa(int(count)),
// 		ev.Name,
// 		asStatus(e.diagnose(ev.Type)),
// 	}

// 	return nil
// }

// func translateMicroTimestampSince(timestamp metav1.MicroTime) string {
// 	if timestamp.IsZero() {
// 		return "<unknown>"
// 	}

// 	return duration.HumanDuration(time.Since(timestamp.Time))
// }

// func translateTimestampSince(timestamp metav1.Time) string {
// 	if timestamp.IsZero() {
// 		return "<unknown>"
// 	}

// 	return duration.HumanDuration(time.Since(timestamp.Time))
// }
// func fmtEventSource(es api.EventSource, reportingController, reportingInstance string) string {
// 	return fmtEventSourceComponentInstance(
// 		firstNonEmpty(es.Component, reportingController),
// 		firstNonEmpty(es.Host, reportingInstance),
// 	)
// }

// func fmtEventSourceComponentInstance(component, instance string) string {
// 	if len(instance) == 0 {
// 		return component
// 	}
// 	return component + ", " + instance
// }

// func firstNonEmpty(ss ...string) string {
// 	for _, s := range ss {
// 		if len(s) > 0 {
// 			return s
// 		}
// 	}
// 	return ""
// }

// // Happy returns true if resource is happy, false otherwise.
// func (Event) diagnose(kind string) error {
// 	if kind != "Normal" {
// 		return errors.New("failed event")
// 	}
// 	return nil
// }

// // Helpers...

// func asRef(r v1.ObjectReference) string {
// 	return strings.ToLower(r.Kind) + ":" + r.Name
// }

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// CronJob renders a K8s CronJob to screen.
type CronJob struct{}

// ColorerFunc colors a resource row.
func (CronJob) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (CronJob) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "SCHEDULE"},
		Header{Name: "SUSPEND"},
		Header{Name: "ACTIVE"},
		Header{Name: "LAST_SCHEDULE"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (c CronJob) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected CronJob, but got %T", o)
	}
	var cj batchv1beta1.CronJob
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cj)
	if err != nil {
		return err
	}

	lastScheduled := "<none>"
	if cj.Status.LastScheduleTime != nil {
		lastScheduled = toAgeHuman(toAge(*cj.Status.LastScheduleTime))
	}

	r.ID = MetaFQN(cj.ObjectMeta)
	r.Fields = make(Fields, 0, len(c.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, cj.Namespace)
	}
	r.Fields = append(r.Fields,
		cj.Name,
		cj.Spec.Schedule,
		boolPtrToStr(cj.Spec.Suspend),
		strconv.Itoa(len(cj.Status.Active)),
		lastScheduled,
		toAge(cj.ObjectMeta.CreationTimestamp),
	)

	return nil
}

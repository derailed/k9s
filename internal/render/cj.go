package render

import (
	"fmt"
	"strconv"

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
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "SCHEDULE"},
		Header{Name: "SUSPEND"},
		Header{Name: "ACTIVE"},
		Header{Name: "LAST_SCHEDULE"},
		Header{Name: "AGE"},
	)
}

// Render renders a K8s resource to screen.
func (CronJob) Render(o interface{}, ns string, r *Row) error {
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

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, cj.Namespace)
	}
	fields = append(fields,
		cj.Name,
		cj.Spec.Schedule,
		boolPtrToStr(cj.Spec.Suspend),
		strconv.Itoa(len(cj.Status.Active)),
		lastScheduled,
		toAge(cj.ObjectMeta.CreationTimestamp),
	)

	r.ID, r.Fields = MetaFQN(cj.ObjectMeta), fields

	return nil
}

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
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
		Header{Name: "SELECTOR", Wide: true},
		Header{Name: "CONTAINERS", Wide: true},
		Header{Name: "IMAGES", Wide: true},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
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

	r.ID = client.MetaFQN(cj.ObjectMeta)
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
		jobSelector(cj.Spec.JobTemplate.Spec),
		podContainerNames(cj.Spec.JobTemplate.Spec.Template.Spec, true),
		podImageNames(cj.Spec.JobTemplate.Spec.Template.Spec, true),
		mapToStr(cj.Labels),
		"",
		toAge(cj.ObjectMeta.CreationTimestamp),
	)

	return nil
}

// Helpers

func jobSelector(spec batchv1.JobSpec) string {
	if spec.Selector == nil {
		return MissingValue
	}
	if len(spec.Selector.MatchLabels) > 0 {
		return mapToStr(spec.Selector.MatchLabels)
	}
	if len(spec.Selector.MatchExpressions) == 0 {
		return ""
	}

	ss := make([]string, 0, len(spec.Selector.MatchExpressions))
	for _, e := range spec.Selector.MatchExpressions {
		ss = append(ss, e.String())
	}

	return strings.Join(ss, " ")
}

func podContainerNames(spec v1.PodSpec, includeInit bool) string {
	cc := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))

	if includeInit {
		for _, c := range spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}
	for _, c := range spec.Containers {
		cc = append(cc, c.Name)
	}

	return strings.Join(cc, ",")
}

func podImageNames(spec v1.PodSpec, includeInit bool) string {
	cc := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))

	if includeInit {
		for _, c := range spec.InitContainers {
			cc = append(cc, c.Image)
		}
	}
	for _, c := range spec.Containers {
		cc = append(cc, c.Image)
	}

	return strings.Join(cc, ",")
}

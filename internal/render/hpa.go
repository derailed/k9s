package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/tview"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// HorizontalPodAutoscaler renders a K8s HorizontalPodAutoscaler to screen.
type HorizontalPodAutoscaler struct{}

// ColorerFunc colors a resource row.
func (HorizontalPodAutoscaler) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (HorizontalPodAutoscaler) Header(ns string) HeaderRow {
	var h HeaderRow
	if isAllNamespace(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "REFERENCE"},
		Header{Name: "TARGETS"},
		Header{Name: "MINPODS", Align: tview.AlignRight},
		Header{Name: "MAXPODS", Align: tview.AlignRight},
		Header{Name: "REPLICAS", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (HorizontalPodAutoscaler) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected HorizontalPodAutoscaler, but got %T", o)
	}
	var hpa autoscalingv1.HorizontalPodAutoscaler
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &hpa)
	if err != nil {
		return err
	}

	fields := make(Fields, 0, len(r.Fields))
	if isAllNamespace(ns) {
		fields = append(fields, hpa.Namespace)
	}
	fields = append(fields,
		hpa.ObjectMeta.Name,
		hpa.Spec.ScaleTargetRef.Name,
		toMetrics(hpa.Spec, hpa.Status),
		strconv.Itoa(int(*hpa.Spec.MinReplicas)),
		strconv.Itoa(int(hpa.Spec.MaxReplicas)),
		strconv.Itoa(int(hpa.Status.CurrentReplicas)),
		toAge(hpa.ObjectMeta.CreationTimestamp),
	)
	r.ID, r.Fields = MetaFQN(hpa.ObjectMeta), fields

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toMetrics(spec autoscalingv1.HorizontalPodAutoscalerSpec, status autoscalingv1.HorizontalPodAutoscalerStatus) string {
	current := "<unknown>"
	if status.CurrentCPUUtilizationPercentage != nil {
		current = strconv.Itoa(int(*status.CurrentCPUUtilizationPercentage)) + "%"
	}

	target := "<unknown>"
	if spec.TargetCPUUtilizationPercentage != nil {
		target = strconv.Itoa(int(*spec.TargetCPUUtilizationPercentage))
	}
	return current + "/" + target + "%"
}

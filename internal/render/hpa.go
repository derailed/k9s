package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
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
func (HorizontalPodAutoscaler) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "REFERENCE"},
		HeaderColumn{Name: "TARGETS%"},
		HeaderColumn{Name: "MINPODS", Align: tview.AlignRight},
		HeaderColumn{Name: "MAXPODS", Align: tview.AlignRight},
		HeaderColumn{Name: "REPLICAS", Align: tview.AlignRight},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (h HorizontalPodAutoscaler) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected HorizontalPodAutoscaler, but got %T", o)
	}

	v := raw.Object["apiVersion"]

	switch v {
	case "autoscaling/v1":
		return h.renderV1(raw, ns, r)
	case "autoscaling/v2beta1":
		return h.renderV2b1(raw, ns, r)
	case "autoscaling/v2beta2":
		return h.renderV2b2(raw, ns, r)
	default:
		return fmt.Errorf("Unhandled HPA version %q", v)
	}
}

func (h HorizontalPodAutoscaler) renderV1(raw *unstructured.Unstructured, _ string, r *Row) error {
	var hpa autoscalingv1.HorizontalPodAutoscaler
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &hpa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(hpa.ObjectMeta)
	r.Fields = Fields{
		hpa.Namespace,
		hpa.ObjectMeta.Name,
		hpa.Spec.ScaleTargetRef.Name,
		toMetricsV1(hpa.Spec, hpa.Status),
		strconv.Itoa(int(*hpa.Spec.MinReplicas)),
		strconv.Itoa(int(hpa.Spec.MaxReplicas)),
		strconv.Itoa(int(hpa.Status.CurrentReplicas)),
		"",
		toAge(hpa.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (h HorizontalPodAutoscaler) renderV2b1(raw *unstructured.Unstructured, _ string, r *Row) error {
	var hpa autoscalingv2beta1.HorizontalPodAutoscaler
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &hpa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(hpa.ObjectMeta)
	r.Fields = Fields{
		hpa.Namespace,
		hpa.ObjectMeta.Name,
		hpa.Spec.ScaleTargetRef.Name,
		toMetricsV2b1(hpa.Spec.Metrics, hpa.Status.CurrentMetrics),
		strconv.Itoa(int(*hpa.Spec.MinReplicas)),
		strconv.Itoa(int(hpa.Spec.MaxReplicas)),
		strconv.Itoa(int(hpa.Status.CurrentReplicas)),
		"",
		toAge(hpa.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (h HorizontalPodAutoscaler) renderV2b2(raw *unstructured.Unstructured, _ string, r *Row) error {
	var hpa autoscalingv2beta2.HorizontalPodAutoscaler
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &hpa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(hpa.ObjectMeta)
	r.Fields = Fields{
		hpa.Namespace,
		hpa.ObjectMeta.Name,
		hpa.Spec.ScaleTargetRef.Name,
		toMetricsV2b2(hpa.Spec.Metrics, hpa.Status.CurrentMetrics),
		strconv.Itoa(int(*hpa.Spec.MinReplicas)),
		strconv.Itoa(int(hpa.Spec.MaxReplicas)),
		strconv.Itoa(int(hpa.Status.CurrentReplicas)),
		"",
		toAge(hpa.ObjectMeta.CreationTimestamp),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toMetricsV1(spec autoscalingv1.HorizontalPodAutoscalerSpec, status autoscalingv1.HorizontalPodAutoscalerStatus) string {
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

func toMetricsV2b1(specs []autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	if len(specs) == 0 {
		return MissingValue
	}

	list, count := []string{}, 0
	for i, spec := range specs {
		list = append(list, checkHPAType(i, spec, statuses))
		count++
	}

	max, more := 2, false
	if count > max {
		list, more = list[:max], true
	}

	ret := strings.Join(list, ", ")
	if more {
		return ret + " + " + strconv.Itoa(count-max) + "more..."
	}

	return ret
}

func toMetricsV2b2(specs []autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	if len(specs) == 0 {
		return MissingValue
	}

	list, max, more, count := []string{}, 2, false, 0
	for i, spec := range specs {
		current := "<unknown>"

		switch spec.Type {
		case autoscalingv2beta2.ExternalMetricSourceType:
			list = append(list, externalMetricsV2b2(i, spec, statuses))
		case autoscalingv2beta2.PodsMetricSourceType:
			if len(statuses) > i && statuses[i].Pods != nil {
				current = statuses[i].Pods.Current.AverageValue.String()
			}
			list = append(list, current+"/"+spec.Pods.Target.AverageValue.String())
		case autoscalingv2beta2.ObjectMetricSourceType:
			if len(statuses) > i && statuses[i].Object != nil {
				current = statuses[i].Object.Current.Value.String()
			}
			list = append(list, current+"/"+spec.Object.Target.Value.String())
		case autoscalingv2beta2.ResourceMetricSourceType:
			list = append(list, resourceMetricsV2b2(i, spec, statuses))
		default:
			list = append(list, "<unknown type>")
		}
		count++
	}

	if count > max {
		list, more = list[:max], true
	}

	ret := strings.Join(list, ", ")
	if more {
		return ret + " + " + strconv.Itoa(count-max) + "more..."
	}

	return ret
}

func checkHPAType(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	switch spec.Type {
	case autoscalingv2beta1.ExternalMetricSourceType:
		return externalMetricsV2b1(i, spec, statuses)
	case autoscalingv2beta1.PodsMetricSourceType:
		if len(statuses) > i && statuses[i].Pods != nil {
			current = statuses[i].Pods.CurrentAverageValue.String()
		}
		return current + "/" + spec.Pods.TargetAverageValue.String()
	case autoscalingv2beta1.ObjectMetricSourceType:
		if len(statuses) > i && statuses[i].Object != nil {
			current = statuses[i].Object.CurrentValue.String()
		}
		return current + "/" + spec.Object.TargetValue.String()
	case autoscalingv2beta1.ResourceMetricSourceType:
		return resourceMetricsV2b1(i, spec, statuses)
	}

	return "<unknown type>"
}

func externalMetricsV2b2(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	current := "<unknown>"

	if spec.External.Target.AverageValue != nil {
		if len(statuses) > i && statuses[i].External != nil && &statuses[i].External.Current.AverageValue != nil {
			current = statuses[i].External.Current.AverageValue.String()
		}
		return current + "/" + spec.External.Target.AverageValue.String() + " (avg)"
	}
	if len(statuses) > i && statuses[i].External != nil {
		current = statuses[i].External.Current.Value.String()
	}

	return current + "/" + spec.External.Target.Value.String()
}

func resourceMetricsV2b2(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	current := "<unknown>"

	if spec.Resource.Target.AverageValue != nil {
		if len(statuses) > i && statuses[i].Resource != nil {
			current = statuses[i].Resource.Current.AverageValue.String()
		}
		return current + "/" + spec.Resource.Target.AverageValue.String()
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.Current.AverageUtilization != nil {
		current = IntToStr(int(*statuses[i].Resource.Current.AverageUtilization))
	}

	target := "<auto>"
	if spec.Resource.Target.AverageUtilization != nil {
		target = IntToStr(int(*spec.Resource.Target.AverageUtilization))
	}

	return current + "/" + target
}

func externalMetricsV2b1(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"
	if spec.External.TargetAverageValue != nil {
		if len(statuses) > i && statuses[i].External != nil && &statuses[i].External.CurrentAverageValue != nil {
			current = statuses[i].External.CurrentAverageValue.String()
		}
		return current + "/" + spec.External.TargetAverageValue.String() + " (avg)"
	}
	if len(statuses) > i && statuses[i].External != nil {
		current = statuses[i].External.CurrentValue.String()
	}

	return current + "/" + spec.External.TargetValue.String()
}

func resourceMetricsV2b1(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	if status := checkTargetMetricsV2b1(i, spec, statuses); status != "" {
		return status
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.CurrentAverageUtilization != nil {
		current = IntToStr(int(*statuses[i].Resource.CurrentAverageUtilization))
	}

	target := "<auto>"
	if spec.Resource.TargetAverageUtilization != nil {
		target = IntToStr(int(*spec.Resource.TargetAverageUtilization))
	}

	return current + "/" + target
}

func checkTargetMetricsV2b1(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	if spec.Resource.TargetAverageValue == nil {
		return ""
	}

	var current string
	if len(statuses) > i && statuses[i].Resource != nil {
		current = statuses[i].Resource.CurrentAverageValue.String()
	}
	return current + "/" + spec.Resource.TargetAverageValue.String()
}

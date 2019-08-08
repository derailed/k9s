package resource

import (
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
)

// HorizontalPodAutoscalerV2Beta1 tracks a kubernetes resource.
type HorizontalPodAutoscalerV2Beta1 struct {
	*Base
	instance *autoscalingv2beta1.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscalerV2Beta1List returns a new resource list.
func NewHorizontalPodAutoscalerV2Beta1List(c Connection, ns string) List {
	return NewList(
		ns,
		"hpa",
		NewHorizontalPodAutoscalerV2Beta1(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewHorizontalPodAutoscalerV2Beta1 instantiates a new HorizontalPodAutoscalerV2Beta1.
func NewHorizontalPodAutoscalerV2Beta1(c Connection) *HorizontalPodAutoscalerV2Beta1 {
	hpa := &HorizontalPodAutoscalerV2Beta1{&Base{Connection: c, Resource: k8s.NewHorizontalPodAutoscalerV2Beta1(c)}, nil}
	hpa.Factory = hpa

	return hpa
}

// New builds a new HorizontalPodAutoscalerV2Beta1 instance from a k8s resource.
func (r *HorizontalPodAutoscalerV2Beta1) New(i interface{}) Columnar {
	c := NewHorizontalPodAutoscalerV2Beta1(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv2beta1.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv2beta1.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown HorizontalPodAutoscalerV2Beta1 type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *HorizontalPodAutoscalerV2Beta1) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa := i.(*autoscalingv2beta1.HorizontalPodAutoscaler)
	hpa.TypeMeta.APIVersion = extractVersion(hpa.Annotations)
	hpa.TypeMeta.Kind = "HorizontalPodAutoscaler"

	return r.marshalObject(hpa)
}

// Header return resource header.
func (*HorizontalPodAutoscalerV2Beta1) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh,
		"NAME",
		"REFERENCE",
		"TARGETS",
		"MINPODS",
		"MAXPODS",
		"REPLICAS",
		"AGE")
}

// Fields retrieves displayable fields.
func (r *HorizontalPodAutoscalerV2Beta1) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.ObjectMeta.Name,
		i.Spec.ScaleTargetRef.Name,
		r.toMetrics(i.Spec.Metrics, i.Status.CurrentMetrics),
		strconv.Itoa(int(*i.Spec.MinReplicas)),
		strconv.Itoa(int(i.Spec.MaxReplicas)),
		strconv.Itoa(int(i.Status.CurrentReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *HorizontalPodAutoscalerV2Beta1) toMetrics(specs []autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	if len(specs) == 0 {
		return "<none>"
	}

	list, count := []string{}, 0
	for i, spec := range specs {
		list = append(list, r.checkHPAType(i, spec, statuses))
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

func (r *HorizontalPodAutoscalerV2Beta1) checkHPAType(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	switch spec.Type {
	case autoscalingv2beta1.ExternalMetricSourceType:
		return r.externalMetrics(i, spec, statuses)
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
		return r.resourceMetrics(i, spec, statuses)
	}

	return "<unknown type>"
}

func (*HorizontalPodAutoscalerV2Beta1) externalMetrics(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
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

func (*HorizontalPodAutoscalerV2Beta1) resourceMetrics(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	if status := checkTargetMetrics(i, spec, statuses); status != "" {
		return status
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.CurrentAverageUtilization != nil {
		current = AsPerc(float64(*statuses[i].Resource.CurrentAverageUtilization))
	}

	target := "<auto>"
	if spec.Resource.TargetAverageUtilization != nil {
		target = AsPerc(float64(*spec.Resource.TargetAverageUtilization))
	}

	return current + "/" + target
}

func checkTargetMetrics(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	if spec.Resource.TargetAverageValue == nil {
		return ""
	}

	var current string
	if len(statuses) > i && statuses[i].Resource != nil {
		current = statuses[i].Resource.CurrentAverageValue.String()
	}
	return current + "/" + spec.Resource.TargetAverageValue.String()
}

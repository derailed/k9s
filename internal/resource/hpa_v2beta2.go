package resource

import (
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
)

// HorizontalPodAutoscaler tracks a kubernetes resource.
type HorizontalPodAutoscaler struct {
	*Base
	instance *autoscalingv2beta2.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscalerList returns a new resource list.
func NewHorizontalPodAutoscalerList(c Connection, ns string) List {
	return NewList(
		ns,
		"hpa",
		NewHorizontalPodAutoscaler(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewHorizontalPodAutoscaler instantiates a new HorizontalPodAutoscaler.
func NewHorizontalPodAutoscaler(c Connection) *HorizontalPodAutoscaler {
	hpa := &HorizontalPodAutoscaler{&Base{Connection: c, Resource: k8s.NewHorizontalPodAutoscalerV2Beta2(c)}, nil}
	hpa.Factory = hpa

	return hpa
}

// New builds a new HorizontalPodAutoscaler instance from a k8s resource.
func (r *HorizontalPodAutoscaler) New(i interface{}) Columnar {
	c := NewHorizontalPodAutoscaler(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown HorizontalPodAutoscaler type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *HorizontalPodAutoscaler) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa := i.(*autoscalingv2beta2.HorizontalPodAutoscaler)
	hpa.TypeMeta.APIVersion = "autoscaling/v2beta2"
	hpa.TypeMeta.Kind = "HorizontalPodAutoscaler"

	return r.marshalObject(hpa)
}

// Header return resource header.
func (*HorizontalPodAutoscaler) Header(ns string) Row {
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
func (r *HorizontalPodAutoscaler) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.ObjectMeta.Name,
		i.Spec.ScaleTargetRef.Name,
		toMetrics(i.Spec.Metrics, i.Status.CurrentMetrics),
		strconv.Itoa(int(*i.Spec.MinReplicas)),
		strconv.Itoa(int(i.Spec.MaxReplicas)),
		strconv.Itoa(int(i.Status.CurrentReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func toMetrics(specs []autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	if len(specs) == 0 {
		return "<none>"
	}

	list, max, more, count := []string{}, 2, false, 0
	for i, spec := range specs {
		current := "<unknown>"

		switch spec.Type {
		case autoscalingv2beta2.ExternalMetricSourceType:
			list = append(list, externalMetrics(i, spec, statuses))
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
			list = append(list, resourceMetrics(i, spec, statuses))
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

func externalMetrics(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
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

func resourceMetrics(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	current := "<unknown>"

	if spec.Resource.Target.AverageValue != nil {
		if len(statuses) > i && statuses[i].Resource != nil {
			current = statuses[i].Resource.Current.AverageValue.String()
		}
		return current + "/" + spec.Resource.Target.AverageValue.String()
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.Current.AverageUtilization != nil {
		current = AsPerc(float64(*statuses[i].Resource.Current.AverageUtilization))
	}

	target := "<auto>"
	if spec.Resource.Target.AverageUtilization != nil {
		target = AsPerc(float64(*spec.Resource.Target.AverageUtilization))
	}

	return current + "/" + target
}

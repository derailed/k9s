package resource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
)

// HPA tracks a kubernetes resource.
type HPA struct {
	*Base
	instance *autoscalingv2beta2.HorizontalPodAutoscaler
}

// NewHPAList returns a new resource list.
func NewHPAList(c Connection, ns string) List {
	return NewList(
		ns,
		"hpa",
		NewHPA(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewHPA instantiates a new HPA.
func NewHPA(c Connection) *HPA {
	hpa := &HPA{&Base{Connection: c, Resource: k8s.NewHPAV2Beta2(c)}, nil}
	hpa.Factory = hpa

	return hpa
}

// New builds a new HPA instance from a k8s resource.
func (r *HPA) New(i interface{}) Columnar {
	c := NewHPA(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown HPA type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *HPA) Marshal(path string) (string, error) {
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
func (*HPA) Header(ns string) Row {
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
func (r *HPA) Fields(ns string) Row {
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
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Pods.Target.AverageValue.String()))
		case autoscalingv2beta2.ObjectMetricSourceType:
			if len(statuses) > i && statuses[i].Object != nil {
				current = statuses[i].Object.Current.Value.String()
			}
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Object.Target.Value.String()))
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
		return fmt.Sprintf("%s + %d more...", ret, count-max)
	}

	return ret
}

func externalMetrics(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	current := "<unknown>"

	if spec.External.Target.AverageValue != nil {
		if len(statuses) > i && statuses[i].External != nil && &statuses[i].External.Current.AverageValue != nil {
			current = statuses[i].External.Current.AverageValue.String()
		}
		return fmt.Sprintf("%s/%s (avg)", current, spec.External.Target.AverageValue.String())
	}
	if len(statuses) > i && statuses[i].External != nil {
		current = statuses[i].External.Current.Value.String()
	}

	return fmt.Sprintf("%s/%s", current, spec.External.Target.Value.String())
}

func resourceMetrics(i int, spec autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	current := "<unknown>"

	if spec.Resource.Target.AverageValue != nil {
		if len(statuses) > i && statuses[i].Resource != nil {
			current = statuses[i].Resource.Current.AverageValue.String()
		}
		return fmt.Sprintf("%s/%s", current, spec.Resource.Target.AverageValue.String())
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.Current.AverageUtilization != nil {
		current = fmt.Sprintf("%d%%", *statuses[i].Resource.Current.AverageUtilization)
	}

	target := "<auto>"
	if spec.Resource.Target.AverageUtilization != nil {
		target = fmt.Sprintf("%d%%", *spec.Resource.Target.AverageUtilization)
	}
	return fmt.Sprintf("%s/%s", current, target)
}

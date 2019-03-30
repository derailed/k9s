package resource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
)

// HPAV2Beta1 tracks a kubernetes resource.
type HPAV2Beta1 struct {
	*Base
	instance *autoscalingv2beta1.HorizontalPodAutoscaler
}

// NewHPAV2Beta1List returns a new resource list.
func NewHPAV2Beta1List(c Connection, ns string) List {
	return NewList(
		ns,
		"hpa",
		NewHPAV2Beta1(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewHPAV2Beta1 instantiates a new HPAV2Beta1.
func NewHPAV2Beta1(c Connection) *HPAV2Beta1 {
	hpa := &HPAV2Beta1{&Base{Connection: c, Resource: k8s.NewHPAV2Beta1(c)}, nil}
	hpa.Factory = hpa

	return hpa
}

// New builds a new HPAV2Beta1 instance from a k8s resource.
func (r *HPAV2Beta1) New(i interface{}) Columnar {
	c := NewHPAV2Beta1(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv2beta1.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv2beta1.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown HPAV2Beta1 type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *HPAV2Beta1) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa := i.(*autoscalingv2beta1.HorizontalPodAutoscaler)
	hpa.TypeMeta.APIVersion = "autoscaling/v2beta1"
	hpa.TypeMeta.Kind = "HorizontalPodAutoscaler"

	return r.marshalObject(hpa)
}

// Header return resource header.
func (*HPAV2Beta1) Header(ns string) Row {
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
func (r *HPAV2Beta1) Fields(ns string) Row {
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

func (r *HPAV2Beta1) toMetrics(specs []autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	if len(specs) == 0 {
		return "<none>"
	}

	list, max, more, count := []string{}, 2, false, 0
	for i, spec := range specs {
		current := "<unknown>"

		switch spec.Type {
		case autoscalingv2beta1.ExternalMetricSourceType:
			list = append(list, r.externalMetrics(i, spec, statuses))
		case autoscalingv2beta1.PodsMetricSourceType:
			if len(statuses) > i && statuses[i].Pods != nil {
				current = statuses[i].Pods.CurrentAverageValue.String()
			}
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Pods.TargetAverageValue.String()))
		case autoscalingv2beta1.ObjectMetricSourceType:
			if len(statuses) > i && statuses[i].Object != nil {
				current = statuses[i].Object.CurrentValue.String()
			}
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Object.TargetValue.String()))
		case autoscalingv2beta1.ResourceMetricSourceType:
			list = append(list, r.resourceMetrics(i, spec, statuses))
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

func (*HPAV2Beta1) externalMetrics(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	if spec.External.TargetAverageValue != nil {
		if len(statuses) > i && statuses[i].External != nil && &statuses[i].External.CurrentAverageValue != nil {
			current = statuses[i].External.CurrentAverageValue.String()
		}
		return fmt.Sprintf("%s/%s (avg)", current, spec.External.TargetAverageValue.String())
	}
	if len(statuses) > i && statuses[i].External != nil {
		current = statuses[i].External.CurrentValue.String()
	}

	return fmt.Sprintf("%s/%s", current, spec.External.TargetValue.String())
}

func (*HPAV2Beta1) resourceMetrics(i int, spec autoscalingv2beta1.MetricSpec, statuses []autoscalingv2beta1.MetricStatus) string {
	current := "<unknown>"

	if spec.Resource.TargetAverageValue != nil {
		if len(statuses) > i && statuses[i].Resource != nil {
			current = statuses[i].Resource.CurrentAverageValue.String()
		}
		return fmt.Sprintf("%s/%s", current, spec.Resource.TargetAverageValue.String())
	}

	if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.CurrentAverageUtilization != nil {
		current = fmt.Sprintf("%d%%", *statuses[i].Resource.CurrentAverageUtilization)
	}

	target := "<auto>"
	if spec.Resource.TargetAverageUtilization != nil {
		target = fmt.Sprintf("%d%%", *spec.Resource.TargetAverageUtilization)
	}
	return fmt.Sprintf("%s/%s", current, target)
}

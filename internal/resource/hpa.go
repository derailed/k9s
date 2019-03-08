package resource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
)

// HPA tracks a kubernetes resource.
type HPA struct {
	*Base
	instance *autoscalingv2beta2.HorizontalPodAutoscaler
}

// NewHPAList returns a new resource list.
func NewHPAList(ns string) List {
	return NewHPAListWithArgs(ns, NewHPA())
}

// NewHPAListWithArgs returns a new resource list.
func NewHPAListWithArgs(ns string, res Resource) List {
	return newList(ns, "hpa", res, AllVerbsAccess|DescribeAccess)
}

// NewHPA instantiates a new Endpoint.
func NewHPA() *HPA {
	return NewHPAWithArgs(k8s.NewHPA())
}

// NewHPAWithArgs instantiates a new Endpoint.
func NewHPAWithArgs(r k8s.Res) *HPA {
	ep := &HPA{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*HPA) NewInstance(i interface{}) Columnar {
	cm := NewHPA()
	switch i.(type) {
	case *autoscalingv2beta2.HorizontalPodAutoscaler:
		cm.instance = i.(*autoscalingv2beta2.HorizontalPodAutoscaler)
	case autoscalingv2beta2.HorizontalPodAutoscaler:
		ii := i.(autoscalingv2beta2.HorizontalPodAutoscaler)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *HPA) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa := i.(*v1.HorizontalPodAutoscaler)
	hpa.TypeMeta.APIVersion = "autoscaling/v1"
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

// ExtFields returns extended fields in relation to headers.
func (*HPA) ExtFields() Properties {
	return Properties{}
}

func toMetrics(specs []autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	if len(specs) == 0 {
		return "<none>"
	}
	list, max, more, count := []string{}, 2, false, 0
	for i, spec := range specs {
		switch spec.Type {
		case autoscalingv2beta2.ExternalMetricSourceType:
			current := "<unknown>"
			if spec.External.Target.AverageValue != nil {
				if len(statuses) > i && statuses[i].External != nil && &statuses[i].External.Current.AverageValue != nil {
					current = statuses[i].External.Current.AverageValue.String()
				}
				list = append(list, fmt.Sprintf("%s/%s (avg)", current, spec.External.Target.AverageValue.String()))
			} else {
				if len(statuses) > i && statuses[i].External != nil {
					current = statuses[i].External.Current.Value.String()
				}
				list = append(list, fmt.Sprintf("%s/%s", current, spec.External.Target.Value.String()))
			}
		case autoscalingv2beta2.PodsMetricSourceType:
			current := "<unknown>"
			if len(statuses) > i && statuses[i].Pods != nil {
				current = statuses[i].Pods.Current.AverageValue.String()
			}
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Pods.Target.AverageValue.String()))
		case autoscalingv2beta2.ObjectMetricSourceType:
			current := "<unknown>"
			if len(statuses) > i && statuses[i].Object != nil {
				current = statuses[i].Object.Current.Value.String()
			}
			list = append(list, fmt.Sprintf("%s/%s", current, spec.Object.Target.Value.String()))
		case autoscalingv2beta2.ResourceMetricSourceType:
			current := "<unknown>"
			if spec.Resource.Target.AverageValue != nil {
				if len(statuses) > i && statuses[i].Resource != nil {
					current = statuses[i].Resource.Current.AverageValue.String()
				}
				list = append(list, fmt.Sprintf("%s/%s", current, spec.Resource.Target.AverageValue.String()))
			} else {
				if len(statuses) > i && statuses[i].Resource != nil && statuses[i].Resource.Current.AverageUtilization != nil {
					current = fmt.Sprintf("%d%%", *statuses[i].Resource.Current.AverageUtilization)
				}
				target := "<auto>"
				if spec.Resource.Target.AverageUtilization != nil {
					target = fmt.Sprintf("%d%%", *spec.Resource.Target.AverageUtilization)
				}
				list = append(list, fmt.Sprintf("%s/%s", current, target))
			}
		default:
			list = append(list, "<unknown type>")
		}
		count++
	}

	if count > max {
		list = list[:max]
		more = true
	}

	ret := strings.Join(list, ", ")
	if more {
		return fmt.Sprintf("%s + %d more...", ret, count-max)
	}
	return ret
}

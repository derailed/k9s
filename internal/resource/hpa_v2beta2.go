package resource

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
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
func (r *HorizontalPodAutoscaler) New(i interface{}) (Columnar, error) {
	c := NewHorizontalPodAutoscaler(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv2beta2.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting HPAv2b2 but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal resource to yaml.
func (r *HorizontalPodAutoscaler) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa, ok := i.(*autoscalingv2beta2.HorizontalPodAutoscaler)
	if !ok {
		return "", errors.New("expecting hpa resource")
	}
	hpa.TypeMeta.APIVersion = extractVersion(hpa.Annotations)
	hpa.TypeMeta.Kind = "HorizontalPodAutoscaler"

	return r.marshalObject(hpa)
}

func extractVersion(a map[string]string) string {
	ann := a["kubectl.kubernetes.io/last-applied-configuration"]
	rx := regexp.MustCompile(`\A{"apiVersion":"([\w|/]+)",`)
	found := rx.FindAllStringSubmatch(ann, 1)
	if len(found) == 0 || len(found[0]) < 1 {
		return "autoscaling/v2beta2"
	}

	return found[0][1]
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

func computePodStatus(ss []autoscalingv2beta2.MetricStatus, index int, current string) string {
	if len(ss) > index && ss[index].Pods != nil {
		return ss[index].Pods.Current.AverageValue.String()
	}
	return current
}

func computeObjectStatus(ss []autoscalingv2beta2.MetricStatus, index int, current string) string {
	if len(ss) > index && ss[index].Object != nil {
		return ss[index].Object.Current.Value.String()
	}
	return current
}

func toMetrics(specs []autoscalingv2beta2.MetricSpec, statuses []autoscalingv2beta2.MetricStatus) string {
	if len(specs) == 0 {
		return MissingValue
	}

	list, max, more, count := []string{}, 2, false, 0
	for i, spec := range specs {
		current := UnknownValue

		switch spec.Type {
		case autoscalingv2beta2.ExternalMetricSourceType:
			list = append(list, externalMetrics(i, spec, statuses))
		case autoscalingv2beta2.PodsMetricSourceType:
			list = append(list, computePodStatus(statuses, i, current)+"/"+spec.Pods.Target.AverageValue.String())
		case autoscalingv2beta2.ObjectMetricSourceType:
			list = append(list, computeObjectStatus(statuses, i, current)+"/"+spec.Object.Target.Value.String())
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
	current := UnknownValue

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
	current := UnknownValue

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

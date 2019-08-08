package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

// HorizontalPodAutoscalerV1 tracks a kubernetes resource.
type HorizontalPodAutoscalerV1 struct {
	*Base
	instance *autoscalingv1.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscalerV1List returns a new resource list.
func NewHorizontalPodAutoscalerV1List(c Connection, ns string) List {
	return NewList(
		ns,
		"hpa",
		NewHorizontalPodAutoscalerV1(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewHorizontalPodAutoscalerV1 instantiates a new HorizontalPodAutoscalerV1.
func NewHorizontalPodAutoscalerV1(c Connection) *HorizontalPodAutoscalerV1 {
	hpa := &HorizontalPodAutoscalerV1{&Base{Connection: c, Resource: k8s.NewHorizontalPodAutoscalerV1(c)}, nil}
	hpa.Factory = hpa

	return hpa
}

// New builds a new HorizontalPodAutoscalerV1 instance from a k8s resource.
func (r *HorizontalPodAutoscalerV1) New(i interface{}) Columnar {
	c := NewHorizontalPodAutoscalerV1(r.Connection)
	switch instance := i.(type) {
	case *autoscalingv1.HorizontalPodAutoscaler:
		c.instance = instance
	case autoscalingv1.HorizontalPodAutoscaler:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown HorizontalPodAutoscalerV1 type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *HorizontalPodAutoscalerV1) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	hpa := i.(*autoscalingv1.HorizontalPodAutoscaler)
	hpa.TypeMeta.APIVersion = extractVersion(hpa.Annotations)
	hpa.TypeMeta.Kind = "HorizontalPodAutoscaler"

	return r.marshalObject(hpa)
}

// Header return resource header.
func (*HorizontalPodAutoscalerV1) Header(ns string) Row {
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
func (r *HorizontalPodAutoscalerV1) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.ObjectMeta.Name,
		i.Spec.ScaleTargetRef.Name,
		r.toMetrics(i.Spec, i.Status),
		strconv.Itoa(int(*i.Spec.MinReplicas)),
		strconv.Itoa(int(i.Spec.MaxReplicas)),
		strconv.Itoa(int(i.Status.CurrentReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *HorizontalPodAutoscalerV1) toMetrics(spec autoscalingv1.HorizontalPodAutoscalerSpec, status autoscalingv1.HorizontalPodAutoscalerStatus) string {
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

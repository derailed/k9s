package resource

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/autoscaling/v1"
)

// HPA tracks a kubernetes resource.
type HPA struct {
	*Base
	instance *v1.HorizontalPodAutoscaler
}

// NewHPAList returns a new resource list.
func NewHPAList(ns string) List {
	return NewHPAListWithArgs(ns, NewHPA())
}

// NewHPAListWithArgs returns a new resource list.
func NewHPAListWithArgs(ns string, res Resource) List {
	return newList(ns, "hpa", res, AllVerbsAccess)
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
	case *v1.HorizontalPodAutoscaler:
		cm.instance = i.(*v1.HorizontalPodAutoscaler)
	case v1.HorizontalPodAutoscaler:
		ii := i.(v1.HorizontalPodAutoscaler)
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
	hpa.TypeMeta.Kind = "HPA"
	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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

	target := "<unknown>"
	if i.Status.CurrentCPUUtilizationPercentage != nil {
		target = strconv.Itoa(int(*i.Status.CurrentCPUUtilizationPercentage))
	}

	var current int32
	if i.Spec.TargetCPUUtilizationPercentage != nil {
		current = *i.Spec.TargetCPUUtilizationPercentage
	}

	return append(ff,
		i.ObjectMeta.Name,
		i.Spec.ScaleTargetRef.Name,
		fmt.Sprintf("%s٪/%d٪", target, current),
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

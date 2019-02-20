package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

// Deployment tracks a kubernetes resource.
type Deployment struct {
	*Base
	instance *v1.Deployment
}

// NewDeploymentList returns a new resource list.
func NewDeploymentList(ns string) List {
	return NewDeploymentListWithArgs(ns, NewDeployment())
}

// NewDeploymentListWithArgs returns a new resource list.
func NewDeploymentListWithArgs(ns string, res Resource) List {
	return newList(ns, "deploy", res, AllVerbsAccess|DescribeAccess)
}

// NewDeployment instantiates a new Deployment.
func NewDeployment() *Deployment {
	return NewDeploymentWithArgs(k8s.NewDeployment())
}

// NewDeploymentWithArgs instantiates a new Deployment.
func NewDeploymentWithArgs(r k8s.Res) *Deployment {
	cm := &Deployment{
		Base: &Base{
			caller: r,
		},
	}
	cm.creator = cm
	return cm
}

// NewInstance builds a new Deployment instance from a k8s resource.
func (*Deployment) NewInstance(i interface{}) Columnar {
	cm := NewDeployment()
	switch i.(type) {
	case *v1.Deployment:
		cm.instance = i.(*v1.Deployment)
	case v1.Deployment:
		ii := i.(v1.Deployment)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *Deployment) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	dp := i.(*v1.Deployment)
	dp.TypeMeta.APIVersion = "apps/v1"
	dp.TypeMeta.Kind = "Deployment"
	return r.marshalObject(dp)
}

// Header return resource header.
func (*Deployment) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE")
}

// Fields retrieves displayable fields.
func (r *Deployment) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}
	return append(ff,
		i.Name,
		strconv.Itoa(int(*i.Spec.Replicas)),
		strconv.Itoa(int(i.Status.Replicas)),
		strconv.Itoa(int(i.Status.UpdatedReplicas)),
		strconv.Itoa(int(i.Status.AvailableReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Deployment) ExtFields() Properties {
	return Properties{}
}

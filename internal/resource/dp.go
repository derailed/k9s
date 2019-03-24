package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
)

// Deployment tracks a kubernetes resource.
type Deployment struct {
	*Base
	instance *v1.Deployment
}

// NewDeploymentList returns a new resource list.
func NewDeploymentList(c k8s.Connection, ns string) List {
	return newList(
		ns,
		"deploy",
		NewDeployment(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewDeployment instantiates a new Deployment.
func NewDeployment(c k8s.Connection) *Deployment {
	d := &Deployment{&Base{connection: c, resource: k8s.NewDeployment(c)}, nil}
	d.Factory = d

	return d
}

// New builds a new Deployment instance from a k8s resource.
func (r *Deployment) New(i interface{}) Columnar {
	c := NewDeployment(r.connection)
	switch instance := i.(type) {
	case *v1.Deployment:
		c.instance = instance
	case v1.Deployment:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Deployment type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Deployment) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.resource.Get(ns, n)
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

package resource

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
)

// Deployment tracks a kubernetes resource.
type Deployment struct {
	*Base
	instance *appsv1.Deployment
}

// NewDeploymentList returns a new resource list.
func NewDeploymentList(c Connection, ns string) List {
	return NewList(
		ns,
		"deploy",
		NewDeployment(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewDeployment instantiates a new Deployment.
func NewDeployment(c Connection) *Deployment {
	d := &Deployment{&Base{Connection: c, Resource: k8s.NewDeployment(c)}, nil}
	d.Factory = d

	return d
}

// New builds a new Deployment instance from a k8s resource.
func (r *Deployment) New(i interface{}) Columnar {
	c := NewDeployment(r.Connection)
	switch instance := i.(type) {
	case *appsv1.Deployment:
		c.instance = instance
	case appsv1.Deployment:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Deployment type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Deployment) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	dp := i.(*appsv1.Deployment)
	dp.TypeMeta.APIVersion = "apps/v1"
	dp.TypeMeta.Kind = "Deployment"

	return r.marshalObject(dp)
}

// Logs tail logs for all pods represented by this deployment.
func (r *Deployment) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	instance, err := r.Resource.Get(opts.Namespace, opts.Name)
	if err != nil {
		return err
	}
	dp := instance.(*appsv1.Deployment)
	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on deployment %s", opts.Name)
	}

	return r.podLogs(ctx, c, dp.Spec.Selector.MatchLabels, opts)
}

// Header return resource header.
func (*Deployment) Header(ns string) Row {
	var hh Row
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh,
		"NAME",
		"DESIRED",
		"CURRENT",
		"UP-TO-DATE",
		"AVAILABLE",
		"AGE",
	)
}

// NumCols designates if column is numerical.
func (*Deployment) NumCols(n string) map[string]bool {
	return map[string]bool{
		"DESIRED":    true,
		"CURRENT":    true,
		"UP-TO-DATE": true,
		"AVAILABLE":  true,
	}
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

// Scale the specified resource.
func (r *Deployment) Scale(ns, n string, replicas int32) error {
	return r.Resource.(Scalable).Scale(ns, n, replicas)
}

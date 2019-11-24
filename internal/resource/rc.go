package resource

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/core/v1"
)

// ReplicationController tracks a kubernetes resource.
type ReplicationController struct {
	*Base
	instance *v1.ReplicationController
}

// NewReplicationControllerList returns a new resource list.
func NewReplicationControllerList(c Connection, ns string) List {
	return NewList(
		ns,
		"rc",
		NewReplicationController(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewReplicationController instantiates a new ReplicationController.
func NewReplicationController(c Connection) *ReplicationController {
	r := &ReplicationController{&Base{Connection: c, Resource: k8s.NewReplicationController(c)}, nil}
	r.Factory = r

	return r
}

// New builds a new ReplicationController instance from a k8s resource.
func (r *ReplicationController) New(i interface{}) (Columnar, error) {
	c := NewReplicationController(r.Connection)
	switch instance := i.(type) {
	case *v1.ReplicationController:
		c.instance = instance
	case v1.ReplicationController:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting RC but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal a deployment given a namespaced name.
func (r *ReplicationController) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	rc, ok := i.(*v1.ReplicationController)
	if !ok {
		return "", errors.New("Expecting a rc resource")
	}
	rc.TypeMeta.APIVersion = "v1"
	rc.TypeMeta.Kind = "ReplicationController"

	return r.marshalObject(rc)
}

// Header return resource header.
func (*ReplicationController) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "DESIRED", "CURRENT", "READY", "AGE")
}

// Fields retrieves displayable fields.
func (r *ReplicationController) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	if ns == AllNamespaces {
		ff = append(ff, r.instance.Namespace)
	}
	i := r.instance

	return append(ff,
		i.Name,
		strconv.Itoa(int(*i.Spec.Replicas)),
		strconv.Itoa(int(i.Status.Replicas)),
		strconv.Itoa(int(i.Status.ReadyReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// Scale the specified resource.
func (r *ReplicationController) Scale(ns, n string, replicas int32) error {
	return r.Resource.(Scalable).Scale(ns, n, replicas)
}

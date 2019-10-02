package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// ReplicationController tracks a kubernetes resource.
type ReplicationController struct {
	*Base
	instance *v1.ReplicationController
}

// NewReplicationControllerList returns a new resource list.
func NewReplicationControllerList(c Connection, ns string, gvr k8s.GVR) List {
	return NewList(
		ns,
		"rc",
		NewReplicationController(c, gvr),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewReplicationController instantiates a new ReplicationController.
func NewReplicationController(c Connection, gvr k8s.GVR) *ReplicationController {
	r := &ReplicationController{&Base{Connection: c, Resource: k8s.NewReplicationController(c, gvr)}, nil}
	r.Factory = r

	return r
}

// New builds a new ReplicationController instance from a k8s resource.
func (r *ReplicationController) New(i interface{}) Columnar {
	c := NewReplicationController(r.Connection, r.GVR())
	switch instance := i.(type) {
	case *v1.ReplicationController:
		c.instance = instance
	case v1.ReplicationController:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown ReplicationController type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal a deployment given a namespaced name.
func (r *ReplicationController) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	rc := i.(*v1.ReplicationController)
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

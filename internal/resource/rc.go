package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

// ReplicationController tracks a kubernetes resource.
type ReplicationController struct {
	*Base
	instance *v1.ReplicationController
}

// NewReplicationControllerList returns a new resource list.
func NewReplicationControllerList(ns string) List {
	return NewReplicationControllerListWithArgs(ns, NewReplicationController())
}

// NewReplicationControllerListWithArgs returns a new resource list.
func NewReplicationControllerListWithArgs(ns string, res Resource) List {
	return newList(ns, "rc", res, AllVerbsAccess)
}

// NewReplicationController instantiates a new Endpoint.
func NewReplicationController() *ReplicationController {
	return NewReplicationControllerWithArgs(k8s.NewReplicationController())
}

// NewReplicationControllerWithArgs instantiates a new Endpoint.
func NewReplicationControllerWithArgs(r k8s.Res) *ReplicationController {
	ep := &ReplicationController{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*ReplicationController) NewInstance(i interface{}) Columnar {
	cm := NewReplicationController()
	switch i.(type) {
	case *v1.ReplicationController:
		cm.instance = i.(*v1.ReplicationController)
	case v1.ReplicationController:
		ii := i.(v1.ReplicationController)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal a deployment given a namespaced name.
func (r *ReplicationController) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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

// ExtFields returns extended fields in relation to headers.
func (*ReplicationController) ExtFields() Properties {
	return Properties{}
}

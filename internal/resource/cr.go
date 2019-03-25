package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/rbac/v1"
)

// ClusterRole tracks a kubernetes resource.
type ClusterRole struct {
	*Base
	instance *v1.ClusterRole
}

// NewClusterRoleList returns a new resource list.
func NewClusterRoleList(c k8s.Connection, ns string) List {
	return NewList(
		NotNamespaced,
		"clusterrole",
		NewClusterRole(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewClusterRole instantiates a new ClusterRole.
func NewClusterRole(c k8s.Connection) *ClusterRole {
	cr := &ClusterRole{&Base{Connection: c, Resource: k8s.NewClusterRole(c)}, nil}
	cr.Factory = cr

	return cr
}

// New builds a new ClusterRole instance from a k8s resource.
func (r *ClusterRole) New(i interface{}) Columnar {
	c := NewClusterRole(r.Connection)
	switch instance := i.(type) {
	case *v1.ClusterRole:
		c.instance = instance
	case v1.ClusterRole:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown context type %#v", i)
	}
	c.path = c.instance.Name

	return c
}

// Marshal resource to yaml.
func (r *ClusterRole) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	cr := i.(*v1.ClusterRole)
	cr.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	cr.TypeMeta.Kind = "ClusterRole"

	return r.marshalObject(cr)
}

// Header return resource header.
func (*ClusterRole) Header(ns string) Row {
	return append(Row{}, "NAME", "AGE")
}

// Fields retrieves displayable fields.
func (r *ClusterRole) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	return append(ff,
		Pad(i.Name, 70),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

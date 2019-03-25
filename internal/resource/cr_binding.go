package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/rbac/v1"
)

// ClusterRoleBinding tracks a kubernetes resource.
type ClusterRoleBinding struct {
	*Base
	instance *v1.ClusterRoleBinding
}

// NewClusterRoleBindingList returns a new resource list.
func NewClusterRoleBindingList(c Connection, _ string) List {
	return NewList(
		NotNamespaced,
		"clusterrolebinding",
		NewClusterRoleBinding(c),
		ViewAccess|DeleteAccess|DescribeAccess,
	)
}

// NewClusterRoleBinding instantiates a new ClusterRoleBinding.
func NewClusterRoleBinding(c Connection) *ClusterRoleBinding {
	crb := &ClusterRoleBinding{&Base{Connection: c, Resource: k8s.NewClusterRoleBinding(c)}, nil}
	crb.Factory = crb

	return crb
}

// New builds a new tabular instance from a k8s resource.
func (r *ClusterRoleBinding) New(i interface{}) Columnar {
	crb := NewClusterRoleBinding(r.Connection)
	switch instance := i.(type) {
	case *v1.ClusterRoleBinding:
		crb.instance = instance
	case v1.ClusterRoleBinding:
		crb.instance = &instance
	default:
		log.Fatal().Msgf("unknown context type %#v", i)
	}
	crb.path = crb.instance.Name

	return crb
}

// Marshal resource to yaml.
func (r *ClusterRoleBinding) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	crb := i.(*v1.ClusterRoleBinding)
	crb.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	crb.TypeMeta.Kind = "ClusterRoleBinding"

	return r.marshalObject(crb)
}

// Header return resource header.
func (*ClusterRoleBinding) Header(_ string) Row {
	return append(Row{}, "NAME", "AGE")
}

// Fields retrieves displayable fields.
func (r *ClusterRoleBinding) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	return append(ff,
		r.instance.Name,
		toAge(r.instance.ObjectMeta.CreationTimestamp),
	)
}

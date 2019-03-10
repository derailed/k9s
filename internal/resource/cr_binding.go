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
func NewClusterRoleBindingList(ns string) List {
	return NewClusterRoleBindingListWithArgs(ns, NewClusterRoleBinding())
}

// NewClusterRoleBindingListWithArgs returns a new resource list.
func NewClusterRoleBindingListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "ctx", res, SwitchAccess|ViewAccess|DeleteAccess|DescribeAccess)
}

// NewClusterRoleBinding instantiates a new ClusterRoleBinding.
func NewClusterRoleBinding() *ClusterRoleBinding {
	return NewClusterRoleBindingWithArgs(k8s.NewClusterRoleBinding())
}

// NewClusterRoleBindingWithArgs instantiates a new Context.
func NewClusterRoleBindingWithArgs(r k8s.Res) *ClusterRoleBinding {
	ctx := &ClusterRoleBinding{
		Base: &Base{
			caller: r,
		},
	}
	ctx.creator = ctx
	return ctx
}

// NewInstance builds a new Context instance from a k8s resource.
func (r *ClusterRoleBinding) NewInstance(i interface{}) Columnar {
	c := NewClusterRoleBinding()
	switch i.(type) {
	case *v1.ClusterRoleBinding:
		c.instance = i.(*v1.ClusterRoleBinding)
	case v1.ClusterRoleBinding:
		ii := i.(v1.ClusterRoleBinding)
		c.instance = &ii
	default:
		log.Fatal().Msgf("unknown context type %#v", i)
	}
	c.path = c.instance.Name
	return c
}

// Marshal resource to yaml.
func (r *ClusterRoleBinding) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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
	i := r.instance

	return append(ff,
		i.Name,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*ClusterRoleBinding) ExtFields() Properties {
	return Properties{}
}

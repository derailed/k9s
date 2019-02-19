package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/rbac/v1"

	"gopkg.in/yaml.v2"
)

// ClusterRole tracks a kubernetes resource.
type ClusterRole struct {
	*Base
	instance *v1.ClusterRole
}

// NewClusterRoleList returns a new resource list.
func NewClusterRoleList(ns string) List {
	return NewClusterRoleListWithArgs(ns, NewClusterRole())
}

// NewClusterRoleListWithArgs returns a new resource list.
func NewClusterRoleListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "clusterrole", res, CRUDAccess)
}

// NewClusterRole instantiates a new ClusterRole.
func NewClusterRole() *ClusterRole {
	return NewClusterRoleWithArgs(k8s.NewClusterRole())
}

// NewClusterRoleWithArgs instantiates a new Context.
func NewClusterRoleWithArgs(r k8s.Res) *ClusterRole {
	ctx := &ClusterRole{
		Base: &Base{
			caller: r,
		},
	}
	ctx.creator = ctx
	return ctx
}

// NewInstance builds a new Context instance from a k8s resource.
func (r *ClusterRole) NewInstance(i interface{}) Columnar {
	c := NewClusterRole()
	switch i.(type) {
	case *v1.ClusterRole:
		c.instance = i.(*v1.ClusterRole)
	case v1.ClusterRole:
		ii := i.(v1.ClusterRole)
		c.instance = &ii
	default:
		log.Fatalf("unknown context type %#v", i)
	}
	c.path = c.instance.Name
	return c
}

// Marshal resource to yaml.
func (r *ClusterRole) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	cr := i.(*v1.ClusterRole)
	cr.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	cr.TypeMeta.Kind = "ClusterRole"
	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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

// ExtFields returns extended fields in relation to headers.
func (*ClusterRole) ExtFields() Properties {
	return Properties{}
}

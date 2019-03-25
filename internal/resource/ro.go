package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/rbac/v1"
)

// Role tracks a kubernetes resource.
type Role struct {
	*Base
	instance *v1.Role
}

// NewRoleList returns a new resource list.
func NewRoleList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"role",
		NewRole(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewRole instantiates a new Role.
func NewRole(c k8s.Connection) *Role {
	r := &Role{&Base{Connection: c, Resource: k8s.NewRole(c)}, nil}
	r.Factory = r

	return r
}

// New builds a new Role instance from a k8s resource.
func (r *Role) New(i interface{}) Columnar {
	c := NewRole(r.Connection)
	switch instance := i.(type) {
	case *v1.Role:
		c.instance = instance
	case v1.Role:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Role type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Role) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	role := i.(*v1.Role)
	role.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	role.TypeMeta.Kind = "Role"

	return r.marshalObject(role)
}

// Header return resource header.
func (*Role) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "AGE")
}

// Fields retrieves displayable fields.
func (r *Role) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *Role) parseRules(pp []v1.PolicyRule) []Row {
	acc := make([]Row, len(pp))

	for i, p := range pp {
		acc[i] = make(Row, 0, 4)
		acc[i] = append(acc[i], strings.Join(p.Resources, ", "))
		acc[i] = append(acc[i], strings.Join(p.NonResourceURLs, ", "))
		acc[i] = append(acc[i], strings.Join(p.ResourceNames, ", "))
		acc[i] = append(acc[i], strings.Join(p.Verbs, ", "))
	}

	return acc
}

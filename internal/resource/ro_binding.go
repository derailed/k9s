package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/rbac/v1"
)

// RoleBinding tracks a kubernetes resource.
type RoleBinding struct {
	*Base
	instance *v1.RoleBinding
}

// NewRoleBindingList returns a new resource list.
func NewRoleBindingList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"rolebinding",
		NewRoleBinding(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewRoleBinding instantiates a new RoleBinding.
func NewRoleBinding(c k8s.Connection) *RoleBinding {
	r := &RoleBinding{&Base{Connection: c, Resource: k8s.NewRoleBinding(c)}, nil}
	r.Factory = r

	return r
}

// New builds a new RoleBinding instance from a k8s resource.
func (r *RoleBinding) New(i interface{}) Columnar {
	c := NewRoleBinding(r.Connection)
	switch instance := i.(type) {
	case *v1.RoleBinding:
		c.instance = instance
	case v1.RoleBinding:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown RoleBinding type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *RoleBinding) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	rb := i.(*v1.RoleBinding)
	rb.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	rb.TypeMeta.Kind = "RoleBinding"

	return r.marshalObject(rb)
}

// Header return resource header.
func (*RoleBinding) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "ROLE", "SUBJECTS", "AGE")
}

// Fields retrieves displayable fields.
func (r *RoleBinding) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		i.RoleRef.Name,
		r.toSubjects(i.Subjects),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *RoleBinding) toSubjects(ss []v1.Subject) string {
	var acc string
	for i, s := range ss {
		acc += s.Name + "/" + r.toSubjectAlias(s.Kind)
		if i < len(ss)-1 {
			acc += ","
		}
	}

	return acc
}

func (r *RoleBinding) toSubjectAlias(s string) string {
	switch s {
	case v1.UserKind:
		return "USR"
	case v1.GroupKind:
		return "GRP"
	case v1.ServiceAccountKind:
		return "SA"
	default:
		return strings.ToUpper(s)
	}
}

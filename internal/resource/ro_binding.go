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
func NewRoleBindingList(ns string) List {
	return NewRoleBindingListWithArgs(ns, NewRoleBinding())
}

// NewRoleBindingListWithArgs returns a new resource list.
func NewRoleBindingListWithArgs(ns string, res Resource) List {
	return newList(ns, "rolebinding", res, AllVerbsAccess|DescribeAccess)
}

// NewRoleBinding instantiates a new Endpoint.
func NewRoleBinding() *RoleBinding {
	return NewRoleBindingWithArgs(k8s.NewRoleBinding())
}

// NewRoleBindingWithArgs instantiates a new Endpoint.
func NewRoleBindingWithArgs(r k8s.Res) *RoleBinding {
	ep := &RoleBinding{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*RoleBinding) NewInstance(i interface{}) Columnar {
	cm := NewRoleBinding()
	switch i.(type) {
	case *v1.RoleBinding:
		cm.instance = i.(*v1.RoleBinding)
	case v1.RoleBinding:
		ii := i.(v1.RoleBinding)
		cm.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *RoleBinding) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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

// ExtFields returns extended fields in relation to headers.
func (*RoleBinding) ExtFields() Properties {
	return Properties{}
}

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

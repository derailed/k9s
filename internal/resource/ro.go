package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/api/rbac/v1"
)

// Role tracks a kubernetes resource.
type Role struct {
	*Base
	instance *v1.Role
}

// NewRoleList returns a new resource list.
func NewRoleList(ns string) List {
	return NewRoleListWithArgs(ns, NewRole())
}

// NewRoleListWithArgs returns a new resource list.
func NewRoleListWithArgs(ns string, res Resource) List {
	l := newList(ns, "role", res, AllVerbsAccess|DescribeAccess)
	l.xray = true
	return l
}

// NewRole instantiates a new Endpoint.
func NewRole() *Role {
	return NewRoleWithArgs(k8s.NewRole())
}

// NewRoleWithArgs instantiates a new Endpoint.
func NewRoleWithArgs(r k8s.Res) *Role {
	ep := &Role{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Role) NewInstance(i interface{}) Columnar {
	cm := NewRole()
	switch i.(type) {
	case *v1.Role:
		cm.instance = i.(*v1.Role)
	case v1.Role:
		ii := i.(v1.Role)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *Role) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	role := i.(*v1.Role)
	role.TypeMeta.APIVersion = "rbac.authorization.k8s.io/v1"
	role.TypeMeta.Kind = "Role"
	raw, err := yaml.Marshal(role)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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

// ExtFields returns extended fields in relation to headers.
func (r *Role) ExtFields() Properties {
	i := r.instance

	return Properties{
		"Headers": Row{"RESOURCES", "NON-RESOURCE URLS", "RESOURCE NAMES", "VERBS"},
		"Rows":    r.parseRules(i.Rules),
	}
}

// Helpers...

func (r *Role) parseRules(pp []v1.PolicyRule) []Row {
	acc := make([]Row, len(pp))
	for i, p := range pp {
		acc[i] = make(Row, 4)
		acc[i][0] = strings.Join(p.Resources, ", ")
		acc[i][1] = strings.Join(p.NonResourceURLs, ", ")
		acc[i][2] = strings.Join(p.ResourceNames, ", ")
		acc[i][3] = strings.Join(p.Verbs, ", ")
	}
	return acc
}

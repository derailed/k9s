package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

// ServiceAccount represents a Kubernetes resource.
type ServiceAccount struct {
	*Base
	instance *v1.ServiceAccount
}

// NewServiceAccountList returns a new resource list.
func NewServiceAccountList(ns string) List {
	return NewServiceAccountListWithArgs(ns, NewServiceAccount())
}

// NewServiceAccountListWithArgs returns a new resource list.
func NewServiceAccountListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "sa", res, CRUDAccess)
}

// NewServiceAccount instantiates a new Endpoint.
func NewServiceAccount() *ServiceAccount {
	return NewServiceAccountWithArgs(k8s.NewServiceAccount())
}

// NewServiceAccountWithArgs instantiates a new Endpoint.
func NewServiceAccountWithArgs(r k8s.Res) *ServiceAccount {
	ep := &ServiceAccount{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*ServiceAccount) NewInstance(i interface{}) Columnar {
	cm := NewServiceAccount()
	switch i.(type) {
	case *v1.ServiceAccount:
		cm.instance = i.(*v1.ServiceAccount)
	case v1.ServiceAccount:
		ii := i.(v1.ServiceAccount)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *ServiceAccount) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	sa := i.(*v1.ServiceAccount)
	sa.TypeMeta.APIVersion = "v1"
	sa.TypeMeta.Kind = "ServiceAccount"
	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// Header return resource header.
func (*ServiceAccount) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "SECRET", "AGE")
}

// Fields retrieves displayable fields.
func (r *ServiceAccount) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		strconv.Itoa(len(i.Secrets)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*ServiceAccount) ExtFields() Properties {
	return Properties{}
}

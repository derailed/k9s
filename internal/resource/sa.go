package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// ServiceAccount represents a Kubernetes resource.
type ServiceAccount struct {
	*Base
	instance *v1.ServiceAccount
}

// NewServiceAccountList returns a new resource list.
func NewServiceAccountList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"sa",
		NewServiceAccount(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewServiceAccount instantiates a new ServiceAccount.
func NewServiceAccount(c k8s.Connection) *ServiceAccount {
	s := &ServiceAccount{&Base{Connection: c, Resource: k8s.NewServiceAccount(c)}, nil}
	s.Factory = s

	return s
}

// New builds a new ServiceAccount instance from a k8s resource.
func (r *ServiceAccount) New(i interface{}) Columnar {
	c := NewServiceAccount(r.Connection)
	switch instance := i.(type) {
	case *v1.ServiceAccount:
		c.instance = instance
	case v1.ServiceAccount:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown ServiceAccount type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *ServiceAccount) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	sa := i.(*v1.ServiceAccount)
	sa.TypeMeta.APIVersion = "v1"
	sa.TypeMeta.Kind = "ServiceAccount"

	return r.marshalObject(sa)
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

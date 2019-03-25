package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Secret tracks a kubernetes resource.
type Secret struct {
	*Base
	instance *v1.Secret
}

// NewSecretList returns a new resource list.
func NewSecretList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"secret",
		NewSecret(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewSecret instantiates a new Secret.
func NewSecret(c k8s.Connection) *Secret {
	s := &Secret{&Base{Connection: c, Resource: k8s.NewSecret(c)}, nil}
	s.Factory = s

	return s
}

// New builds a new Secret instance from a k8s resource.
func (r *Secret) New(i interface{}) Columnar {
	c := NewSecret(r.Connection)
	switch instance := i.(type) {
	case *v1.Secret:
		c.instance = instance
	case v1.Secret:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Secret type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Secret) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	sec := i.(*v1.Secret)
	sec.TypeMeta.APIVersion = "v1"
	sec.TypeMeta.Kind = "Secret"

	return r.marshalObject(sec)
}

// Header return resource header.
func (*Secret) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "TYPE", "DATA", "AGE")
}

// Fields retrieves displayable fields.
func (r *Secret) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		string(i.Type),
		strconv.Itoa(len(i.Data)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

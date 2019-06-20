package resource

import (
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Endpoints tracks a kubernetes resource.
type Endpoints struct {
	*Base
	instance *v1.Endpoints
}

// NewEndpointsList returns a new resource list.
func NewEndpointsList(c Connection, ns string) List {
	return NewList(
		ns,
		"ep",
		NewEndpoints(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewEndpoints instantiates a new Endpoints.
func NewEndpoints(c Connection) *Endpoints {
	ep := &Endpoints{&Base{Connection: c, Resource: k8s.NewEndpoints(c)}, nil}
	ep.Factory = ep

	return ep
}

// New builds a new Endpoints instance from a k8s resource.
func (r *Endpoints) New(i interface{}) Columnar {
	c := NewEndpoints(r.Connection)
	switch instance := i.(type) {
	case *v1.Endpoints:
		c.instance = instance
	case v1.Endpoints:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Endpoints type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Endpoints) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ep := i.(*v1.Endpoints)
	ep.TypeMeta.APIVersion = "v1"
	ep.TypeMeta.Kind = "Endpoint"

	return r.marshalObject(ep)
}

// Header return resource header.
func (*Endpoints) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "ENDPOINTS", "AGE")
}

// Fields retrieves displayable fields.
func (r *Endpoints) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		missing(r.toEPs(i.Subsets)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *Endpoints) toEPs(ss []v1.EndpointSubset) string {
	aa := make([]string, 0, len(ss))
	max := 3
	for _, s := range ss {
		pp := make([]string, 0, len(s.Ports))
		for _, p := range s.Ports {
			pp = append(pp, strconv.Itoa(int(p.Port)))
		}

		for _, a := range s.Addresses {
			if len(a.IP) == 0 {
				continue
			}
			if len(pp) == 0 {
				aa = append(aa, a.IP)
				continue
			}
			add := a.IP + ":" + strings.Join(pp, ",")
			if len(pp) > max {
				add = a.IP + ":" + strings.Join(pp[:max], ",") + "..."
			}
			aa = append(aa, add)
		}
	}

	return strings.Join(aa, ",")
}

package resource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

// Endpoints tracks a kubernetes resource.
type Endpoints struct {
	*Base
	instance *v1.Endpoints
}

// NewEndpointsList returns a new resource list.
func NewEndpointsList(ns string) List {
	return NewEndpointsListWithArgs(ns, NewEndpoints())
}

// NewEndpointsListWithArgs returns a new resource list.
func NewEndpointsListWithArgs(ns string, res Resource) List {
	return newList(ns, "ep", res, AllVerbsAccess)
}

// NewEndpoints instantiates a new Endpoint.
func NewEndpoints() *Endpoints {
	return NewEndpointsWithArgs(k8s.NewEndpoints())
}

// NewEndpointsWithArgs instantiates a new Endpoint.
func NewEndpointsWithArgs(r k8s.Res) *Endpoints {
	ep := &Endpoints{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Endpoints) NewInstance(i interface{}) Columnar {
	cm := NewEndpoints()
	switch i.(type) {
	case *v1.Endpoints:
		cm.instance = i.(*v1.Endpoints)
	case v1.Endpoints:
		ii := i.(v1.Endpoints)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *Endpoints) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	ep := i.(*v1.Endpoints)
	ep.TypeMeta.APIVersion = "v1"
	ep.TypeMeta.Kind = "Endpoints"
	raw, err := yaml.Marshal(ep)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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

// ExtFields returns extended fields in relation to headers.
func (*Endpoints) ExtFields() Properties {
	return Properties{}
}

func (r *Endpoints) toEPs(ss []v1.EndpointSubset) string {
	aa := make([]string, 0, len(ss))
	max := 3
	for _, s := range ss {
		pp := make([]string, 0, len(s.Ports))
		for _, p := range s.Ports {
			pp = append(pp, strconv.Itoa(int(p.Port)))
		}

		for _, a := range s.Addresses {
			if len(a.IP) != 0 {
				if len(pp) == 0 {
					aa = append(aa, fmt.Sprintf("%s", a.IP))
				} else {
					add := fmt.Sprintf("%s:%s", a.IP, strings.Join(pp, ","))
					if len(pp) > max {
						add = fmt.Sprintf("%s:%s...", a.IP, strings.Join(pp[:max], ","))
					}
					aa = append(aa, add)
				}
			}
		}
	}
	return strings.Join(aa, ",")
}

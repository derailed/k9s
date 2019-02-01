package resource

import (
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/resource/k8s"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

// Service tracks a kubernetes resource.
type Service struct {
	*Base
	instance *v1.Service
}

// NewServiceList returns a new resource list.
func NewServiceList(ns string) List {
	return NewServiceListWithArgs(ns, NewService())
}

// NewServiceListWithArgs returns a new resource list.
func NewServiceListWithArgs(ns string, res Resource) List {
	return newList(ns, "svc", res, AllVerbsAccess)
}

// NewService instantiates a new Endpoint.
func NewService() *Service {
	return NewServiceWithArgs(k8s.NewService())
}

// NewServiceWithArgs instantiates a new Endpoint.
func NewServiceWithArgs(r k8s.Res) *Service {
	ep := &Service{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Service) NewInstance(i interface{}) Columnar {
	cm := NewService()
	switch i.(type) {
	case *v1.Service:
		cm.instance = i.(*v1.Service)
	case v1.Service:
		ii := i.(v1.Service)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
// BOZO!! Why you need to fill type info??
func (r *Service) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	svc := i.(*v1.Service)
	svc.TypeMeta.APIVersion = "v1"
	svc.TypeMeta.Kind = "Service"
	raw, err := yaml.Marshal(svc)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// Header returns resource header.
func (*Service) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh,
		"NAME",
		"TYPE",
		"CLUSTER-IP",
		"EXTERNAL-IP",
		"SELECTOR",
		"PORT(S)",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (r *Service) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.ObjectMeta.Name,
		string(i.Spec.Type),
		i.Spec.ClusterIP,
		r.toIPs(i.Spec.Type, i.Spec.ExternalIPs),
		mapToStr(i.Spec.Selector),
		r.toPorts(i.Spec.Ports),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (r *Service) ExtFields() Properties {
	return Properties{}
}

// Helpers...

func (*Service) toIPs(svcType v1.ServiceType, ips []string) string {
	if len(ips) == 0 {
		if svcType == v1.ServiceTypeLoadBalancer {
			return "<pending>"
		}
		return MissingValue
	}
	sort.Strings(ips)
	return strings.Join(ips, ",")
}

func (*Service) toPorts(pp []v1.ServicePort) string {
	ports := make([]string, len(pp))
	for i, p := range pp {
		if len(p.Name) > 0 {
			ports[i] = p.Name + ":"
		}
		ports[i] += strconv.Itoa(int(p.Port)) +
			"->" +
			strconv.Itoa(int(p.NodePort)) +
			"/" +
			string(p.Protocol)

	}
	return strings.Join(ports, ",")
}

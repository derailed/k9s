package resource

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Service tracks a kubernetes resource.
type Service struct {
	*Base
	instance *v1.Service
}

// NewServiceList returns a new resource list.
func NewServiceList(c Connection, ns string) List {
	return NewList(
		ns,
		"svc",
		NewService(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewService instantiates a new Service.
func NewService(c Connection) *Service {
	s := &Service{&Base{Connection: c, Resource: k8s.NewService(c)}, nil}
	s.Factory = s

	return s
}

// New builds a new Service instance from a k8s resource.
func (r *Service) New(i interface{}) (Columnar, error) {
	c := NewService(r.Connection)
	switch instance := i.(type) {
	case *v1.Service:
		c.instance = instance
	case v1.Service:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting Service but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal resource to yaml.
// BOZO!! Why you need to fill type info??
func (r *Service) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	svc, ok := i.(*v1.Service)
	if !ok {
		return "", errors.New("Expecting a service resource")
	}
	svc.TypeMeta.APIVersion = "v1"
	svc.TypeMeta.Kind = "Service"

	return r.marshalObject(svc)
}

// Logs tail logs for all pods represented by this service.
func (r *Service) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	instance, err := r.Resource.Get(opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	svc, ok := instance.(*v1.Service)
	if !ok {
		return errors.New("Expecting a service resource")
	}
	log.Debug().Msgf("Service %s--%s", svc.Name, svc.Spec.Selector)
	if len(svc.Spec.Selector) == 0 {
		return errors.New("No logs for headless service")
	}

	return r.podLogs(ctx, c, svc.Spec.Selector, opts)
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
		"PORTS",
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
		r.toIPs(i.Spec.Type, r.getSvcExtIPS(i)),
		mapToStr(i.Spec.Selector),
		r.toPorts(i.Spec.Ports),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *Service) getSvcExtIPS(svc *v1.Service) []string {
	results := []string{}

	switch svc.Spec.Type {
	case v1.ServiceTypeClusterIP:
		fallthrough
	case v1.ServiceTypeNodePort:
		return svc.Spec.ExternalIPs
	case v1.ServiceTypeLoadBalancer:
		lbIps := r.lbIngressIP(svc.Status.LoadBalancer)
		if len(svc.Spec.ExternalIPs) > 0 {
			if len(lbIps) > 0 {
				results = append(results, lbIps)
			}
			return append(results, svc.Spec.ExternalIPs...)
		}
		if len(lbIps) > 0 {
			results = append(results, lbIps)
		}
	case v1.ServiceTypeExternalName:
		results = append(results, svc.Spec.ExternalName)
	}

	return results
}

func (*Service) lbIngressIP(s v1.LoadBalancerStatus) string {
	ingress := s.Ingress
	result := []string{}
	for i := range ingress {
		if len(ingress[i].IP) > 0 {
			result = append(result, ingress[i].IP)
		} else if len(ingress[i].Hostname) > 0 {
			result = append(result, ingress[i].Hostname)
		}
	}

	return strings.Join(result, ",")
}

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
			"►" +
			strconv.Itoa(int(p.NodePort))
		if p.Protocol != "TCP" {
			ports[i] += "╱" + string(p.Protocol)
		}
	}

	return strings.Join(ports, " ")
}

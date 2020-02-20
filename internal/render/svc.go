package render

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Service renders a K8s Service to screen.
type Service struct{}

// ColorerFunc colors a resource row.
func (Service) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Service) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "TYPE"},
		Header{Name: "CLUSTER-IP"},
		Header{Name: "EXTERNAL-IP"},
		Header{Name: "SELECTOR", Wide: true},
		Header{Name: "PORTS", Wide: true},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s Service) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Service, but got %T", o)
	}
	var svc v1.Service
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &svc)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(svc.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, svc.Namespace)
	}
	r.Fields = append(r.Fields,
		svc.ObjectMeta.Name,
		string(svc.Spec.Type),
		toIP(svc.Spec.ClusterIP),
		toIPs(svc.Spec.Type, getSvcExtIPS(&svc)),
		mapToStr(svc.Spec.Selector),
		toPorts(svc.Spec.Ports),
		mapToStr(svc.Labels),
		asStatus(s.diagnose()),
		toAge(svc.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (Service) diagnose() error {
	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toIP(ip string) string {
	if ip == "" || ip == "None" {
		return ""
	}
	return ip
}

func getSvcExtIPS(svc *v1.Service) []string {
	results := []string{}

	switch svc.Spec.Type {
	case v1.ServiceTypeClusterIP:
		fallthrough
	case v1.ServiceTypeNodePort:
		return svc.Spec.ExternalIPs
	case v1.ServiceTypeLoadBalancer:
		lbIps := lbIngressIP(svc.Status.LoadBalancer)
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

func lbIngressIP(s v1.LoadBalancerStatus) string {
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

func toIPs(svcType v1.ServiceType, ips []string) string {
	if len(ips) == 0 {
		if svcType == v1.ServiceTypeLoadBalancer {
			return "<pending>"
		}
		return ""
	}
	sort.Strings(ips)

	return strings.Join(ips, ",")
}

func toPorts(pp []v1.ServicePort) string {
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

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Ingress renders a K8s Ingress to screen.
type Ingress struct{}

// ColorerFunc colors a resource row.
func (Ingress) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Ingress) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "CLASS"},
		HeaderColumn{Name: "HOSTS"},
		HeaderColumn{Name: "ADDRESS"},
		HeaderColumn{Name: "PORTS"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (i Ingress) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Ingress, but got %T", o)
	}
	var ing netv1.Ingress
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ing)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(ing.ObjectMeta)
	r.Fields = Fields{
		ing.Namespace,
		ing.Name,
		strpToStr(ing.Spec.IngressClassName),
		toHosts(ing.Spec.Rules),
		toAddress(ing.Status.LoadBalancer),
		toTLSPorts(ing.Spec.TLS),
		mapToStr(ing.Labels),
		toAge(ing.ObjectMeta.CreationTimestamp),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toAddress(lbs v1.LoadBalancerStatus) string {
	ings := lbs.Ingress
	res := make([]string, 0, len(ings))
	for _, lb := range ings {
		if len(lb.IP) > 0 {
			res = append(res, lb.IP)
		} else if len(lb.Hostname) != 0 {
			res = append(res, lb.Hostname)
		}
	}

	return strings.Join(res, ",")
}

func toTLSPorts(tls []netv1.IngressTLS) string {
	if len(tls) != 0 {
		return "80, 443"
	}

	return "80"
}

func toHosts(rr []netv1.IngressRule) string {
	hh := make([]string, 0, len(rr))
	for _, r := range rr {
		if r.Host == "" {
			r.Host = "*"
		}
		hh = append(hh, r.Host)
	}

	return strings.Join(hh, ",")
}

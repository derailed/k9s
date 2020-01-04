package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
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
func (Ingress) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "HOSTS"},
		Header{Name: "ADDRESS"},
		Header{Name: "PORT"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (i Ingress) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Ingress, but got %T", o)
	}
	var ing v1beta1.Ingress
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ing)
	if err != nil {
		return err
	}

	r.ID = MetaFQN(ing.ObjectMeta)
	r.Fields = make(Fields, 0, len(i.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, ing.Namespace)
	}
	r.Fields = append(r.Fields,
		ing.Name,
		toHosts(ing.Spec.Rules),
		toAddress(ing.Status.LoadBalancer),
		toTLSPorts(ing.Spec.TLS),
		toAge(ing.ObjectMeta.CreationTimestamp),
	)

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

func toTLSPorts(tls []v1beta1.IngressTLS) string {
	if len(tls) != 0 {
		return "80, 443"
	}

	return "80"
}

func toHosts(rr []v1beta1.IngressRule) string {
	hh := make([]string, 0, len(rr))
	for _, r := range rr {
		if r.Host == "" {
			r.Host = "*"
		}
		hh = append(hh, r.Host)
	}

	return strings.Join(hh, ",")
}

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Endpoints renders a K8s Endpoints to screen.
type Endpoints struct{}

// ColorerFunc colors a resource row.
func (Endpoints) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Endpoints) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "ENDPOINTS"},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (e Endpoints) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Endpoints, but got %T", o)
	}
	var ep v1.Endpoints
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ep)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(ep.ObjectMeta)
	r.Fields = make(Fields, 0, len(e.Header(ns)))
	r.Fields = Fields{
		ep.Namespace,
		ep.Name,
		missing(toEPs(ep.Subsets)),
		toAge(ep.ObjectMeta.CreationTimestamp),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toEPs(ss []v1.EndpointSubset) string {
	aa := make([]string, 0, len(ss))
	for _, s := range ss {
		pp := make([]string, len(s.Ports))
		portsToStrs(s.Ports, pp)
		a := make([]string, len(s.Addresses))
		proccessIPs(a, pp, s.Addresses)
		aa = append(aa, strings.Join(a, ","))
	}
	return strings.Join(aa, ",")
}

func portsToStrs(pp []v1.EndpointPort, ss []string) {
	for i := 0; i < len(pp); i++ {
		ss[i] = strconv.Itoa(int(pp[i].Port))
	}
}

func proccessIPs(aa []string, pp []string, addrs []v1.EndpointAddress) {
	const maxIPs = 3
	var i int
	for _, a := range addrs {
		if len(a.IP) == 0 {
			continue
		}
		if len(pp) == 0 {
			aa[i], i = a.IP, i+1
			continue
		}
		if len(pp) > maxIPs {
			aa[i], i = a.IP+":"+strings.Join(pp[:maxIPs], ",")+"...", i+1
		} else {
			aa[i], i = a.IP+":"+strings.Join(pp, ","), i+1
		}
	}
}

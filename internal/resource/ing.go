package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

// Ingress tracks a kubernetes resource.
type Ingress struct {
	*Base
	instance *v1beta1.Ingress
}

// NewIngressList returns a new resource list.
func NewIngressList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"ing",
		NewIngress(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewIngress instantiates a new Ingress.
func NewIngress(c k8s.Connection) *Ingress {
	ing := &Ingress{&Base{Connection: c, Resource: k8s.NewIngress(c)}, nil}
	ing.Factory = ing

	return ing
}

// New builds a new Ingress instance from a k8s resource.
func (r *Ingress) New(i interface{}) Columnar {
	c := NewIngress(r.Connection)
	switch instance := i.(type) {
	case *v1beta1.Ingress:
		c.instance = instance
	case v1beta1.Ingress:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Ingress type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Ingress) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ing := i.(*v1beta1.Ingress)
	ing.TypeMeta.APIVersion = "extensions/v1beta1"
	ing.TypeMeta.Kind = "Ingress"

	return r.marshalObject(ing)
}

// Header return resource header.
func (*Ingress) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "HOSTS", "ADDRESS", "PORT", "AGE")
}

// Fields retrieves displayable fields.
func (r *Ingress) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		r.toHosts(i.Spec.Rules),
		r.toAddress(i.Status.LoadBalancer),
		r.toPorts(i.Spec.TLS),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (*Ingress) toAddress(lbs v1.LoadBalancerStatus) string {
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

func (*Ingress) toPorts(tls []v1beta1.IngressTLS) string {
	if len(tls) != 0 {
		return "80, 443"
	}

	return "80"
}

func (*Ingress) toHosts(rr []v1beta1.IngressRule) string {
	var s string
	var i int
	for _, r := range rr {
		s += r.Host
		if i < len(rr)-1 {
			s += ","
		}
		i++
	}

	return s
}

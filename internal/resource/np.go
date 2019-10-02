package resource

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicy tracks a kubernetes resource.
type NetworkPolicy struct {
	*Base
	instance *v1beta1.NetworkPolicy
}

// NewNetworkPolicyList returns a new resource list.
func NewNetworkPolicyList(c Connection, ns string, gvr k8s.GVR) List {
	return NewList(
		ns,
		"np",
		NewNetworkPolicy(c, gvr),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewNetworkPolicy instantiates a new NetworkPolicy.
func NewNetworkPolicy(c Connection, gvr k8s.GVR) *NetworkPolicy {
	ds := &NetworkPolicy{&Base{Connection: c, Resource: k8s.NewNetworkPolicy(c, gvr)}, nil}
	ds.Factory = ds

	return ds
}

// New builds a new NetworkPolicy instance from a k8s resource.
func (r *NetworkPolicy) New(i interface{}) Columnar {
	c := NewNetworkPolicy(r.Connection, r.GVR())
	switch instance := i.(type) {
	case *v1beta1.NetworkPolicy:
		c.instance = instance
	case v1beta1.NetworkPolicy:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown NetworkPolicy type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *NetworkPolicy) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ds := i.(*v1beta1.NetworkPolicy)
	ds.TypeMeta.APIVersion = "extensions/v1beta1"
	ds.TypeMeta.Kind = "NetworkPolicy"

	return r.marshalObject(ds)
}

// Header return resource header.
func (*NetworkPolicy) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	hh = append(hh, "NAME", "ING-SELECTOR", "ING-PORTS", "ING-BLOCK", "EGR-SELECTOR", "EGR-PORTS", "EGR-BLOCK", "AGE")

	return hh
}

// Fields retrieves displayable fields.
func (r *NetworkPolicy) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	ip, is, ib := ingress(i.Spec.Ingress)
	ep, es, eb := egress(i.Spec.Egress)

	return append(ff,
		i.Name,
		is,
		ip,
		ib,
		es,
		ep,
		eb,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// Helpers...

func ingress(ii []v1beta1.NetworkPolicyIngressRule) (string, string, string) {
	var ports, sels, blocks []string
	for _, i := range ii {
		if p := portsToStr(i.Ports); p != "" {
			ports = append(ports, p)
		}
		ll, pp := peersToStr(i.From)
		if ll != "" {
			sels = append(sels, ll)
		}
		if pp != "" {
			blocks = append(blocks, pp)
		}
	}
	return strings.Join(ports, ","), strings.Join(sels, ","), strings.Join(blocks, ",")
}

func egress(ee []v1beta1.NetworkPolicyEgressRule) (string, string, string) {
	var ports, sels, blocks []string
	for _, e := range ee {
		if p := portsToStr(e.Ports); p != "" {
			ports = append(ports, p)
		}
		ll, pp := peersToStr(e.To)
		if ll != "" {
			sels = append(sels, ll)
		}
		if pp != "" {
			blocks = append(blocks, pp)
		}
	}
	return strings.Join(ports, ","), strings.Join(sels, ","), strings.Join(blocks, ",")
}

func portsToStr(pp []v1beta1.NetworkPolicyPort) string {
	ports := make([]string, 0, len(pp))
	for _, p := range pp {
		ports = append(ports, string(*p.Protocol)+":"+p.Port.String())
	}
	return strings.Join(ports, ",")
}

func peersToStr(pp []v1beta1.NetworkPolicyPeer) (string, string) {
	sels := make([]string, 0, len(pp))
	ips := make([]string, 0, len(pp))
	for _, p := range pp {
		if peer := renderPeer(p); peer != "" {
			sels = append(sels, peer)
		}

		if p.IPBlock == nil {
			continue
		}
		if b := renderBlock(p.IPBlock); b != "" {
			ips = append(ips, b)
		}
	}
	return strings.Join(sels, ","), strings.Join(ips, ",")
}

func renderBlock(b *v1beta1.IPBlock) string {
	s := b.CIDR

	if len(b.Except) == 0 {
		return s
	}

	e, more := b.Except, false
	if len(b.Except) > 2 {
		e, more = e[:2], true
	}
	if more {
		return s + "[" + strings.Join(e, ",") + "...]"
	}
	return s + "[" + strings.Join(b.Except, ",") + "]"
}

func renderPeer(i v1beta1.NetworkPolicyPeer) string {
	var s string

	if i.PodSelector != nil {
		if m := mapToStr(i.PodSelector.MatchLabels); m != "" {
			s += "po:" + m
		}
		if e := expToStr(i.PodSelector.MatchExpressions); e != "" {
			s += "--" + e
		}
	}

	if i.NamespaceSelector != nil {
		if m := mapToStr(i.NamespaceSelector.MatchLabels); m != "" {
			s += "ns:" + m
		}
		if e := expToStr(i.NamespaceSelector.MatchExpressions); e != "" {
			s += "--" + e
		}
	}

	return s
}

func expToStr(ee []metav1.LabelSelectorRequirement) string {
	ss := make([]string, len(ee))
	for i, e := range ee {
		ss[i] = labToStr(e)
	}
	return strings.Join(ss, ",")
}

func labToStr(e metav1.LabelSelectorRequirement) string {
	return fmt.Sprintf("%s-%s%s", e.Key, e.Operator, strings.Join(e.Values, ","))
}

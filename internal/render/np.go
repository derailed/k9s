package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// NetworkPolicy renders a K8s NetworkPolicy to screen.
type NetworkPolicy struct{}

// ColorerFunc colors a resource row.
func (NetworkPolicy) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (NetworkPolicy) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "ING-SELECTOR"},
		Header{Name: "ING-PORTS"},
		Header{Name: "ING-BLOCK"},
		Header{Name: "EGR-SELECTOR"},
		Header{Name: "EGR-PORTS"},
		Header{Name: "EGR-BLOCK"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (n NetworkPolicy) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected NetworkPolicy, but got %T", o)
	}
	var np v1beta1.NetworkPolicy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &np)
	if err != nil {
		return err
	}

	ip, is, ib := ingress(np.Spec.Ingress)
	ep, es, eb := egress(np.Spec.Egress)

	r.ID = client.MetaFQN(np.ObjectMeta)
	r.Fields = make(Fields, 0, len(n.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, np.Namespace)
	}
	r.Fields = append(r.Fields,
		np.Name,
		is,
		ip,
		ib,
		es,
		ep,
		eb,
		toAge(np.ObjectMeta.CreationTimestamp),
	)

	return nil
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
		proto, port := NAValue, NAValue
		if p.Protocol != nil {
			proto = string(*p.Protocol)
		}
		if p.Port != nil {
			port = p.Port.String()
		}
		ports = append(ports, proto+":"+port)
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

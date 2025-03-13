// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// NetworkPolicy renders a K8s NetworkPolicy to screen.
type NetworkPolicy struct {
	Base
}

// Header returns a header row.
func (p NetworkPolicy) Header(_ string) model1.Header {
	return p.doHeader(p.defaultHeader())
}

func (NetworkPolicy) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "POD-SELECTOR"},
		model1.HeaderColumn{Name: "ING-SELECTOR", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "ING-PORTS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "ING-BLOCK", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "EGR-SELECTOR", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "EGR-PORTS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "EGR-BLOCK", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (p NetworkPolicy) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected NetworkPolicy, but got %T", o)
	}
	if err := p.defaultRow(raw, row); err != nil {
		return err
	}
	if p.specs.isEmpty() {
		return nil
	}

	cols, err := p.specs.realize(raw, p.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (n NetworkPolicy) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var np netv1.NetworkPolicy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &np)
	if err != nil {
		return err
	}

	ip, is, ib := ingress(np.Spec.Ingress)
	ep, es, eb := egress(np.Spec.Egress)

	var podSel string
	if len(np.Spec.PodSelector.MatchLabels) > 0 {
		podSel = mapToStr(np.Spec.PodSelector.MatchLabels)
	}
	if len(np.Spec.PodSelector.MatchExpressions) > 0 {
		podSel += "::" + expToStr(np.Spec.PodSelector.MatchExpressions)
	}
	r.ID = client.MetaFQN(np.ObjectMeta)
	r.Fields = model1.Fields{
		np.Namespace,
		np.Name,
		podSel,
		is,
		ip,
		ib,
		es,
		ep,
		eb,
		mapToStr(np.Labels),
		"",
		ToAge(np.GetCreationTimestamp()),
	}

	return nil
}

// Helpers...

func ingress(ii []netv1.NetworkPolicyIngressRule) (string, string, string) {
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

func egress(ee []netv1.NetworkPolicyEgressRule) (string, string, string) {
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

func portsToStr(pp []netv1.NetworkPolicyPort) string {
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

func peersToStr(pp []netv1.NetworkPolicyPeer) (string, string) {
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

func renderBlock(b *netv1.IPBlock) string {
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

func renderPeer(i netv1.NetworkPolicyPeer) string {
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

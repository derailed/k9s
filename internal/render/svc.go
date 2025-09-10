// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Header returns a header row.
var defaultSVCHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "TYPE"},
	model1.HeaderColumn{Name: "CLUSTER-IP"},
	model1.HeaderColumn{Name: "EXTERNAL-IP"},
	model1.HeaderColumn{Name: "SELECTOR", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "PORTS", Attrs: model1.Attrs{Wide: false}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Service renders a K8s Service to screen.
type Service struct {
	Base
}

// Header returns a header row.
func (s Service) Header(_ string) model1.Header {
	return s.doHeader(defaultSVCHeader)
}

// Render renders a K8s resource to screen.
func (s Service) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := s.defaultRow(raw, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}
	cols, err := s.specs.realize(raw, defaultSVCHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (s Service) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var svc v1.Service
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &svc)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(&svc.ObjectMeta)
	r.Fields = model1.Fields{
		svc.Namespace,
		svc.Name,
		string(svc.Spec.Type),
		toIP(svc.Spec.ClusterIP),
		toIPs(svc.Spec.Type, getSvcExtIPS(&svc)),
		mapToStr(svc.Spec.Selector),
		ToPorts(svc.Spec.Ports),
		mapToStr(svc.Labels),
		AsStatus(s.diagnose()),
		ToAge(svc.GetCreationTimestamp()),
	}

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
	case v1.ServiceTypeNodePort, v1.ServiceTypeClusterIP:
		return svc.Spec.ExternalIPs
	case v1.ServiceTypeLoadBalancer:
		lbIps := lbIngressIP(svc.Status.LoadBalancer)
		if len(svc.Spec.ExternalIPs) > 0 {
			if lbIps != "" {
				results = append(results, lbIps)
			}
			return append(results, svc.Spec.ExternalIPs...)
		}
		if lbIps != "" {
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
		if ingress[i].IP != "" {
			result = append(result, ingress[i].IP)
		} else if ingress[i].Hostname != "" {
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

// ToPorts returns service ports as a string.
func ToPorts(pp []v1.ServicePort) string {
	ports := make([]string, len(pp))
	for i, p := range pp {
		if p.Name != "" {
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

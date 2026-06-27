// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
)

var defaultGatewayAPIHeader = model1.Header{
	model1.HeaderColumn{Name: "RESOURCE_TYPE"},
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "AGE"},
}

// GatewayAPIRes represents a Gateway API resource.
type GatewayAPIRes struct {
	metav1.TableRow
}

// GetObjectKind returns the object kind.
func (*GatewayAPIRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a deep copy.
func (g *GatewayAPIRes) DeepCopyObject() runtime.Object {
	return nil
}

// GatewayAPI represents a Gateway API renderer.
type GatewayAPI struct {
	Base
}

// Header returns the header row.
func (GatewayAPI) Header(_ string) model1.Header {
	return defaultGatewayAPIHeader
}

// Render renders a Gateway API resource to a row.
func (GatewayAPI) Render(o any, _ string, row *model1.Row) error {
	res, ok := o.(*GatewayAPIRes)
	if !ok {
		return fmt.Errorf("expected GatewayAPIRes, but got %T", o)
	}

	if len(res.Cells) < 5 {
		return fmt.Errorf("invalid row cells length: %d", len(res.Cells))
	}

	gvrStr := res.Cells[0].(string)
	namespace := res.Cells[1].(string)
	name := res.Cells[2].(string)
	status := res.Cells[3].(string)
	ts := res.Cells[4].(metav1.Time)

	age := "-"
	if !ts.IsZero() {
		age = duration.HumanDuration(time.Since(ts.Time))
	}

	resourceType := formatResourceType(gvrStr)

	row.ID = fmt.Sprintf("%s|%s|%s", gvrStr, namespace, name)
	row.Fields = model1.Fields{
		resourceType,
		namespace,
		name,
		status,
		age,
	}

	return nil
}

func formatResourceType(gvr string) string {
	switch gvr {
	case "gateway.networking.k8s.io/v1/gateways":
		return "Gateway"
	case "gateway.networking.k8s.io/v1/gatewayclasses":
		return "GatewayClass"
	case "gateway.networking.k8s.io/v1/httproutes":
		return "HTTPRoute"
	case "gateway.networking.k8s.io/v1/grpcroutes":
		return "GRPCRoute"
	case "gateway.networking.k8s.io/v1/tcproutes":
		return "TCPRoute"
	case "gateway.networking.k8s.io/v1/udproutes":
		return "UDPRoute"
	case "gateway.networking.k8s.io/v1/tlsroutes":
		return "TLSRoute"
	default:
		return gvr
	}
}

// ColorerFunc colors a resource row.
func (GatewayAPI) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("STATUS", true)
		if !ok {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[idx])
		if status == "DEGRADED" {
			c = model1.PendingColor
		}

		return c
	}
}
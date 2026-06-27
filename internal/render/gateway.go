// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultGatewayHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "CLASS"},
	model1.HeaderColumn{Name: "ADDRESSES"},
	model1.HeaderColumn{Name: "PORTS"},
	model1.HeaderColumn{Name: "READY"},
	model1.HeaderColumn{Name: "AGE"},
	model1.HeaderColumn{Name: "STATUS"},
}

type Gateway struct {
	Base
}

func (g Gateway) Header(_ string) model1.Header {
	return g.doHeader(defaultGatewayHeader)
}

func (g Gateway) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := g.defaultRow(raw, row); err != nil {
		return err
	}
	if g.specs.isEmpty() {
		return nil
	}
	cols, err := g.specs.realize(raw, defaultGatewayHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (g Gateway) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta, _ := raw.Object["metadata"].(map[string]any)
	spec, _ := raw.Object["spec"].(map[string]any)
	status, _ := raw.Object["status"].(map[string]any)
	if meta == nil {
		return fmt.Errorf("missing metadata in Gateway resource")
	}

	namespace, _ := meta["namespace"].(string)
	name, _ := meta["name"].(string)

	className := ""
	if classRef, ok := spec["gatewayClassName"].(string); ok {
		className = classRef
	}

	addresses := formatGatewayAddresses(status["addresses"])
	ports := formatGatewayPorts(spec["listeners"])
	ready := getGatewayReadyStatus(status)
	age := getTimestampAge(meta["creationTimestamp"])
	statusMsg := getStatusMessage(status["conditions"], "Ready")

	r.ID = client.FQN(namespace, name)
	r.Fields = model1.Fields{
		namespace,
		name,
		className,
		addresses,
		ports,
		ready,
		age,
		statusMsg,
	}

	return nil
}

func (g Gateway) diagnose() error {
	return nil
}
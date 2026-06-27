// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultGatewayClassHeader = model1.Header{
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "CONTROLLER"},
	model1.HeaderColumn{Name: "AGE"},
	model1.HeaderColumn{Name: "STATUS"},
}

type GatewayClass struct {
	Base
}

func (gc GatewayClass) Header(_ string) model1.Header {
	return gc.doHeader(defaultGatewayClassHeader)
}

func (gc GatewayClass) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := gc.defaultRow(raw, row); err != nil {
		return err
	}
	if gc.specs.isEmpty() {
		return nil
	}
	cols, err := gc.specs.realize(raw, defaultGatewayClassHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (gc GatewayClass) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta, _ := raw.Object["metadata"].(map[string]any)
	spec, _ := raw.Object["spec"].(map[string]any)
	status, _ := raw.Object["status"].(map[string]any)
	if meta == nil {
		return fmt.Errorf("missing metadata in GatewayClass resource")
	}

	namespace, _ := meta["namespace"].(string)
	name, _ := meta["name"].(string)

	controller := ""
	if controllerName, ok := spec["controllerName"].(string); ok {
		controller = controllerName
	}

	age := getTimestampAge(meta["creationTimestamp"])
	statusMsg := getStatusMessage(status["conditions"], "Accepted")

	if namespace == "" {
		r.ID = name
	} else {
		r.ID = client.FQN(namespace, name)
	}

	r.Fields = model1.Fields{
		name,
		controller,
		age,
		statusMsg,
	}

	return nil
}

func (gc GatewayClass) diagnose() error {
	return nil
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultTCPRouteHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "SERVICES"},
	model1.HeaderColumn{Name: "PARENTS"},
	model1.HeaderColumn{Name: "AGE"},
	model1.HeaderColumn{Name: "STATUS"},
}

type TCPRoute struct {
	Base
}

func (tr TCPRoute) Header(_ string) model1.Header {
	return tr.doHeader(defaultTCPRouteHeader)
}

func (tr TCPRoute) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := tr.defaultRow(raw, row); err != nil {
		return err
	}
	if tr.specs.isEmpty() {
		return nil
	}
	cols, err := tr.specs.realize(raw, defaultTCPRouteHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (tr TCPRoute) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta, _ := raw.Object["metadata"].(map[string]any)
	spec, _ := raw.Object["spec"].(map[string]any)
	status, _ := raw.Object["status"].(map[string]any)
	if meta == nil {
		return fmt.Errorf("missing metadata in TCPRoute resource")
	}

	namespace, _ := meta["namespace"].(string)
	name, _ := meta["name"].(string)

	services := formatRouteServices(spec["rules"])
	parents := formatRouteParents(status["parents"])
	age := getTimestampAge(meta["creationTimestamp"])
	statusMsg := getStatusMessage(status["parents"], "Accepted")

	r.ID = client.FQN(namespace, name)
	r.Fields = model1.Fields{
		namespace,
		name,
		services,
		parents,
		age,
		statusMsg,
	}

	return nil
}

func (tr TCPRoute) diagnose() error {
	return nil
}
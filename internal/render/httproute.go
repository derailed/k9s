// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultHTTPRouteHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "HOSTNAMES"},
	model1.HeaderColumn{Name: "SERVICES"},
	model1.HeaderColumn{Name: "PARENTS"},
	model1.HeaderColumn{Name: "AGE"},
	model1.HeaderColumn{Name: "STATUS"},
}

type HTTPRoute struct {
	Base
}

func (hr HTTPRoute) Header(_ string) model1.Header {
	return hr.doHeader(defaultHTTPRouteHeader)
}

func (hr HTTPRoute) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := hr.defaultRow(raw, row); err != nil {
		return err
	}
	if hr.specs.isEmpty() {
		return nil
	}
	cols, err := hr.specs.realize(raw, defaultHTTPRouteHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (hr HTTPRoute) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta, _ := raw.Object["metadata"].(map[string]any)
	spec, _ := raw.Object["spec"].(map[string]any)
	status, _ := raw.Object["status"].(map[string]any)
	if meta == nil {
		return fmt.Errorf("missing metadata in HTTPRoute resource")
	}

	namespace, _ := meta["namespace"].(string)
	name, _ := meta["name"].(string)

	hostnames := formatRouteHostnames(spec["hostnames"])
	services := formatRouteServices(spec["rules"])
	parents := formatRouteParents(status["parents"])
	age := getTimestampAge(meta["creationTimestamp"])
	statusMsg := getStatusMessage(status["parents"], "Accepted")

	r.ID = client.FQN(namespace, name)
	r.Fields = model1.Fields{
		namespace,
		name,
		hostnames,
		services,
		parents,
		age,
		statusMsg,
	}

	return nil
}

func (hr HTTPRoute) diagnose() error {
	return nil
}
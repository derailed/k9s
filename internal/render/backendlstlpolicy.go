// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultBackendTLSPolicyHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "TARGETS"},
	model1.HeaderColumn{Name: "AGE"},
	model1.HeaderColumn{Name: "STATUS"},
}

type BackendTLSPolicy struct {
	Base
}

func (btp BackendTLSPolicy) Header(_ string) model1.Header {
	return btp.doHeader(defaultBackendTLSPolicyHeader)
}

func (btp BackendTLSPolicy) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := btp.defaultRow(raw, row); err != nil {
		return err
	}
	if btp.specs.isEmpty() {
		return nil
	}
	cols, err := btp.specs.realize(raw, defaultBackendTLSPolicyHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (btp BackendTLSPolicy) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta := raw.Object["metadata"].(map[string]any)
	spec := raw.Object["spec"].(map[string]any)
	status := raw.Object["status"].(map[string]any)

	namespace := meta["namespace"].(string)
	name := meta["name"].(string)

	targets := formatBackendTLSPolicyTargets(spec["targetRefs"])
	age := getTimestampAge(meta["creationTimestamp"])
	statusMsg := getStatusMessage(status["conditions"], "Accepted")

	r.ID = client.FQN(namespace, name)
	r.Fields = model1.Fields{
		namespace,
		name,
		targets,
		age,
		statusMsg,
	}

	return nil
}

func (btp BackendTLSPolicy) diagnose() error {
	return nil
}

func formatBackendTLSPolicyTargets(targets any) string {
	result := ""
	if targetList, ok := targets.([]any); ok {
		var targetItems []string
		for _, t := range targetList {
			if target, ok := t.(map[string]any); ok {
				group, groupOk := target["group"].(string)
				kind, kindOk := target["kind"].(string)
				name, nameOk := target["name"].(string)
				namespace, namespaceOk := target["namespace"].(string)

				if groupOk && kindOk && nameOk {
					if namespaceOk && namespace != "" {
						targetItems = append(targetItems, fmt.Sprintf("%s/%s.%s/%s", group, kind, namespace, name))
					} else {
						targetItems = append(targetItems, fmt.Sprintf("%s/%s/%s", group, kind, name))
					}
				}
			}
		}
		result = joinStrings(targetItems)
		if len(targetItems) > 2 {
			result = fmt.Sprintf("%s (+%d)", targetItems[0], len(targetItems)-1)
		}
	}
	return result
}
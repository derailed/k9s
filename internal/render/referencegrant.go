// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultReferenceGrantHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "FROM"},
	model1.HeaderColumn{Name: "TO"},
	model1.HeaderColumn{Name: "AGE"},
}

type ReferenceGrant struct {
	Base
}

func (rg ReferenceGrant) Header(_ string) model1.Header {
	return rg.doHeader(defaultReferenceGrantHeader)
}

func (rg ReferenceGrant) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := rg.defaultRow(raw, row); err != nil {
		return err
	}
	if rg.specs.isEmpty() {
		return nil
	}
	cols, err := rg.specs.realize(raw, defaultReferenceGrantHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (rg ReferenceGrant) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	meta := raw.Object["metadata"].(map[string]any)
	spec := raw.Object["spec"].(map[string]any)

	namespace := meta["namespace"].(string)
	name := meta["name"].(string)

	from := formatReferenceGrantFrom(spec["from"])
	to := formatReferenceGrantTo(spec["to"])
	age := getTimestampAge(meta["creationTimestamp"])

	r.ID = client.FQN(namespace, name)
	r.Fields = model1.Fields{
		namespace,
		name,
		from,
		to,
		age,
	}

	return nil
}

func (rg ReferenceGrant) diagnose() error {
	return nil
}

func formatReferenceGrantFrom(from any) string {
	result := ""
	if fromList, ok := from.([]any); ok {
		var fromItems []string
		for _, f := range fromList {
			if fromItem, ok := f.(map[string]any); ok {
				group, groupOk := fromItem["group"].(string)
				kind, kindOk := fromItem["kind"].(string)
				namespace, namespaceOk := fromItem["namespace"].(string)

				if groupOk && kindOk {
					if namespaceOk && namespace != "" {
						fromItems = append(fromItems, fmt.Sprintf("%s/%s.%s", group, kind, namespace))
					} else {
						fromItems = append(fromItems, fmt.Sprintf("%s/%s", group, kind))
					}
				}
			}
		}
		result = joinStrings(fromItems)
		if len(fromItems) > 2 {
			result = fmt.Sprintf("%s (+%d)", fromItems[0], len(fromItems)-1)
		}
	}
	return result
}

func formatReferenceGrantTo(to any) string {
	result := ""
	if toList, ok := to.([]any); ok {
		var toItems []string
		for _, t := range toList {
			if toItem, ok := t.(map[string]any); ok {
				group, groupOk := toItem["group"].(string)
				kind, kindOk := toItem["kind"].(string)

				if groupOk && kindOk {
					toItems = append(toItems, fmt.Sprintf("%s/%s", group, kind))
				}
			}
		}
		result = joinStrings(toItems)
		if len(toItems) > 2 {
			result = fmt.Sprintf("%s (+%d)", toItems[0], len(toItems)-1)
		}
	}
	return result
}
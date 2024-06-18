package dao

import (
	"fmt"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/util/jsonpath"
	"reflect"
	"strings"
)

// extractPathFromObject extracts a json path from a given object.
func extractPathFromObject(o runtime.Object, path string) ([]string, error) {
	var err error
	parser := jsonpath.New("accessor").AllowMissingKeys(true)
	parser.EnableJSONOutput(true)
	fullPath := fmt.Sprintf("{%s}", path)
	if err := parser.Parse(fullPath); err != nil {
		return nil, err
	}
	log.Debug().Msgf("Prepared JSONPath %s.", fullPath)

	var results [][]reflect.Value
	if unstructured, ok := o.(runtime.Unstructured); ok {
		results, err = parser.FindResults(unstructured.UnstructuredContent())
	} else {
		results, err = parser.FindResults(reflect.ValueOf(o).Elem().Interface())
	}

	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Results extracted %s.", results)

	if len(results) != 1 {
		return nil, nil
	}

	if err != nil {
		return nil, nil
	}
	var values = make([]string, 0)
	for arrIx := range results {
		for valIx := range results[arrIx] {
			values = append(values, fmt.Sprint(results[arrIx][valIx].Interface()))
		}
	}
	return values, nil
}

// SelectorsForLink builds label and field selectors from a given object based on a custom resource link.
func SelectorsForLink(link *config.CustomResourceLink, o runtime.Object) (labels.Selector, fields.Selector, error) {
	var labelSelector = labels.Everything()
	var fieldSelector fields.Selector

	for targetLabel, sourceField := range link.LabelSelector {
		values, err := extractPathFromObject(o, sourceField)
		if err != nil {
			return nil, nil, err
		}
		if values == nil {
			continue
		}
		if len(values) != 1 {
			continue
		}
		log.Debug().Msgf("Extracted values for label selector %s: %+v.", targetLabel, values)

		req, err := labels.NewRequirement(targetLabel, selection.Equals, values)
		if err != nil {
			return nil, nil, err
		}
		labelSelector = labelSelector.Add(*req)
	}

	for targetField, sourceField := range link.FieldSelector {
		values, err := extractPathFromObject(o, sourceField)
		if err != nil {
			return nil, nil, err
		}
		if values == nil {
			continue
		}
		log.Debug().Msgf("Extracted values for field selector %s: %+v.", targetField, values)

		sel := fields.OneTermEqualSelector(targetField, strings.Join(values, ","))
		if fieldSelector == nil {
			fieldSelector = sel
			continue
		}
		fieldSelector = fields.AndSelectors(
			fieldSelector,
			sel,
		)
	}
	return labelSelector, fieldSelector, nil
}

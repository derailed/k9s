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

	var values = make([]string, 0)
	for arrIx := range results {
		for valIx := range results[arrIx] {
			values = append(values, fmt.Sprint(results[arrIx][valIx].Interface()))
		}
	}
	return values, nil
}

// SelectorsForLink builds label and field selectors from a given object based on a custom resource link.
func SelectorsForLink(link *config.CustomResourceLink, o runtime.Object) (string, string, error) {
	var labelSelector = labels.Everything()
	var fieldSelector fields.Selector

	for target, source := range link.LabelSelector {
		values, err := extractPathFromObject(o, source)
		switch {
		case err != nil:
			return "", "", err
		case values == nil || len(values) != 1:
			continue
		}
		log.Debug().Msgf("Extracted values for label selector %s: %+v.", target, values)

		req, err := labels.NewRequirement(target, selection.Equals, values)
		if err != nil {
			return "", "", err
		}
		labelSelector = labelSelector.Add(*req)
	}

	for target, source := range link.FieldSelector {
		values, err := extractPathFromObject(o, source)
		switch {
		case err != nil:
			return "", "", err
		case values == nil:
			continue
		}
		log.Debug().Msgf("Extracted values for field selector %s: %+v.", target, values)

		sel := fields.OneTermEqualSelector(target, strings.Join(values, ","))
		if fieldSelector == nil {
			fieldSelector = sel
			continue
		}
		fieldSelector = fields.AndSelectors(
			fieldSelector,
			sel,
		)
	}
	fsel := ""
	if fieldSelector != nil {
		fsel = fieldSelector.String()
	}

	return labelSelector.String(), fsel, nil
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"errors"
	"math"
	"regexp"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

const defaultServiceAccount = "default"

var (
	inverseRx = regexp.MustCompile(`\A\!`)
	fuzzyRx   = regexp.MustCompile(`\A-f\s?([\w-]+)\b`)
)

func inList(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
}

// IsInverseSelector checks if inverse char has been provided.
func IsInverseSelector(s string) bool {
	if s == "" {
		return false
	}
	return inverseRx.MatchString(s)
}

// HasFuzzySelector checks if query is fuzzy.
func HasFuzzySelector(s string) (string, bool) {
	mm := fuzzyRx.FindStringSubmatch(s)
	if len(mm) != 2 {
		return "", false
	}

	return mm[1], true
}

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}

// ToYAML converts a resource to its YAML representation.
func ToYAML(o runtime.Object, showManaged bool) (string, error) {
	if o == nil {
		return "", errors.New("no object to yamlize")
	}

	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	if !showManaged {
		o = o.DeepCopyObject()
		uo := o.(*unstructured.Unstructured).Object
		if meta, ok := uo["metadata"].(map[string]interface{}); ok {
			delete(meta, "managedFields")
		}
	}
	err := p.PrintObj(o, &buff)
	if err != nil {
		log.Error().Msgf("Marshal Error %v", err)
		return "", err
	}

	return buff.String(), nil
}

// serviceAccountMatches validates that the ServiceAccount referenced in the PodSpec matches the incoming
// ServiceAccount. If the PodSpec ServiceAccount is blank kubernetes will use the "default" ServiceAccount
// when deploying the pod, so if the incoming SA is "default" and podSA is an empty string that is also a match.
func serviceAccountMatches(podSA, saName string) bool {
	if podSA == "" {
		podSA = defaultServiceAccount
	}
	return podSA == saName
}

// ContinuousRanges takes a sorted slice of integers and returns a slice of
// sub-slices representing continuous ranges of integers.
func ContinuousRanges(indexes []int) [][]int {
	var ranges [][]int
	for i, p := 1, 0; i <= len(indexes); i++ {
		if i == len(indexes) || indexes[i]-indexes[p] != i-p {
			ranges = append(ranges, []int{indexes[p], indexes[i-1] + 1})
			p = i
		}
	}

	return ranges
}

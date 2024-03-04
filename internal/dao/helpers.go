// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

const (
	defaultServiceAccount      = "default"
	defaultContainerAnnotation = "kubectl.kubernetes.io/default-container"
)

// GetDefaultContainer returns a container name if specified in an annotation.
func GetDefaultContainer(m metav1.ObjectMeta, spec v1.PodSpec) (string, bool) {
	defaultContainer, ok := m.Annotations[defaultContainerAnnotation]
	if !ok {
		return "", false
	}

	for _, container := range spec.Containers {
		if container.Name == defaultContainer {
			return defaultContainer, true
		}
	}
	log.Warn().Msg(defaultContainer + " container  not found. " + defaultContainerAnnotation + " annotation will be ignored")

	return "", false
}

func extractFQN(o runtime.Object) string {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting unstructured but got %T", o))
		return client.NA
	}

	return FQN(u.GetNamespace(), u.GetName())
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

func inList(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
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

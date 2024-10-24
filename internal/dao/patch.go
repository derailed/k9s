// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"encoding/json"
)

// ImageSpec represents a container image.
type ImageSpec struct {
	Index             int
	Name, DockerImage string
	Init              bool
}

// ImageSpecs represents a collection of container images.
type ImageSpecs []ImageSpec

// JsonPatch track pod spec updates.
type JsonPatch struct {
	Spec Spec `json:"spec"`
}

// Spec represents a pod template.
type Spec struct {
	Template PodSpec `json:"template"`
}

// PodSpec represents a collection of container images.
type PodSpec struct {
	Spec ImagesSpec `json:"spec"`
}

// ImagesSpec tracks container image updates.
type ImagesSpec struct {
	SetElementOrderContainers     []Element `json:"$setElementOrder/containers,omitempty"`
	SetElementOrderInitContainers []Element `json:"$setElementOrder/initContainers,omitempty"`
	Containers                    []Element `json:"containers,omitempty"`
	InitContainers                []Element `json:"initContainers,omitempty"`
}

// Element tracks a given container image.
type Element struct {
	Image string `json:"image,omitempty"`
	Name  string `json:"name"`
}

// GetTemplateJsonPatch builds a json patch string to update PodSpec images.
func GetTemplateJsonPatch(imageSpecs ImageSpecs) ([]byte, error) {
	jsonPatch := JsonPatch{
		Spec: Spec{
			Template: getPatchPodSpec(imageSpecs),
		},
	}
	return json.Marshal(jsonPatch)
}

// GetJsonPatch returns container image patch.
func GetJsonPatch(imageSpecs ImageSpecs) ([]byte, error) {
	podSpec := getPatchPodSpec(imageSpecs)
	return json.Marshal(podSpec)
}

func getPatchPodSpec(imageSpecs ImageSpecs) PodSpec {
	initElementsOrders, initElements, elementsOrders, elements := extractElements(imageSpecs)
	podSpec := PodSpec{
		Spec: ImagesSpec{
			SetElementOrderInitContainers: initElementsOrders,
			InitContainers:                initElements,
			SetElementOrderContainers:     elementsOrders,
			Containers:                    elements,
		},
	}
	return podSpec
}

func extractElements(imageSpecs ImageSpecs) (initElementsOrders []Element, initElements []Element, elementsOrders []Element, elements []Element) {
	for _, spec := range imageSpecs {
		if spec.Init {
			initElementsOrders = append(initElementsOrders, Element{Name: spec.Name})
			initElements = append(initElements, Element{Name: spec.Name, Image: spec.DockerImage})
		} else {
			elementsOrders = append(elementsOrders, Element{Name: spec.Name})
			elements = append(elements, Element{Name: spec.Name, Image: spec.DockerImage})
		}
	}
	return initElementsOrders, initElements, elementsOrders, elements
}
